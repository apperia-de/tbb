// Package tbb provides a base for creating custom Telegram bots.
// It can be used to spin up a custom bot in one minute which is capable of
// handling bot users via an implemented sqlite database,
// but can easily be switched to use mysql or postgres instead.
package tbb

import (
	"context"
	"fmt"
	"github.com/NicoNex/echotron/v3"
	"github.com/apperia-de/tbb/pkg/model"
	timezone "github.com/evanoberholster/timezoneLookup/v2"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type TBB struct {
	db     *DB
	dsp    *echotron.Dispatcher
	ctx    context.Context
	cfg    *Config
	logger *slog.Logger
	cmdReg CommandRegistry
	hFn    UpdateHandlerFn
	api    echotron.API // Telegram api
	tzc    *timezone.Timezonecache
	srv    *http.Server
}

type Option func(*TBB)

// New creates a new Telegram bot based on the given configuration.
// It uses functional options for configuration.
func New(opts ...Option) *TBB {
	app := &TBB{
		ctx:    context.Background(),
		cmdReg: CommandRegistry{},
		hFn:    func() UpdateHandler { return &DefaultUpdateHandler{} },
		logger: nil,
		tzc:    loadTimezoneCache(),
	}

	// Loop through each option
	for _, opt := range opts {
		opt(app)
	}

	if app.cfg == nil {
		panic("app config is missing")
	}

	if app.logger == nil {
		app.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     getLogLevel(app.cfg.LogLevel),
		}))
	}

	app.db = NewDB(app.cfg, &gorm.Config{FullSaveAssociations: true})
	app.api = echotron.NewAPI(app.cfg.Telegram.BotToken)
	app.dsp = echotron.NewDispatcher(app.cfg.Telegram.BotToken, app.buildBot(app.hFn))
	if app.srv != nil {
		app.dsp.SetHTTPServer(app.srv)
	}

	// Initialize database tables
	if err := app.db.AutoMigrate(&model.User{}, &model.UserInfo{}, &model.UserPhoto{}); err != nil {
		panic(err)
	}

	return app
}

// WithConfig is the only required option because it provides the config for the app to function properly.
func WithConfig(cfg *Config) Option {
	return func(app *TBB) {
		app.cfg = cfg
	}
}

// WithCommands is used for providing and registering custom bot commands.
// Bot commands always start with a / like /start and a Handler, which implements the CommandHandler interface.
// If you want a command to be available in the command list on Telegram,
// the provided Command must contain a Description.
func WithCommands(commands []Command) Option {
	return func(app *TBB) {
		app.cmdReg = buildCommandRegistry(commands)
	}
}

// WithHandlerFunc option can be used to override the default UpdateHandlerFn for custom echotron.Update message handling.
func WithHandlerFunc(hFn UpdateHandlerFn) Option {
	return func(app *TBB) {
		app.hFn = hFn
	}
}

// WithLogger option can be used to override the default logger with a custom one.
func WithLogger(l *slog.Logger) Option {
	return func(app *TBB) {
		app.logger = l
	}
}

// WithServer option can be used add a custom http.Server to the dispatcher
func WithServer(s *http.Server) Option {
	return func(app *TBB) {
		app.srv = s
	}
}

// Start starts the Telegram bot server in poll mode
func (tb *TBB) Start() {
	if err := tb.SetBotCommands(tb.buildTelegramCommands()); err != nil {
		tb.logger.Error("Cannot set bot commands!")
		panic(err)
	}

	if tb.srv == nil {
		tb.logger.Info("Start dispatcher")
		tb.logger.Error(tb.dsp.Poll().Error())
		return
	}

	// If we have a custom web server, we run the polling in a separate go routine.
	tb.logger.Info("Start dispatcher and web server")
	go tb.logger.Error(tb.dsp.Poll().Error())
	tb.logger.Error(tb.srv.ListenAndServe().Error())
}

// StartWithWebhook starts the Telegram bot server with a given webhook url.
func (tb *TBB) StartWithWebhook(webhookURL string) {
	if err := tb.SetBotCommands(tb.buildTelegramCommands()); err != nil {
		tb.logger.Error("Cannot set bot commands!")
		panic(err)
	}
	if webhookURL == "" {
		panic("webhook url is empty")
	}

	tb.logger.Info(fmt.Sprintf("Start dispatcher with webhook: %q", webhookURL))
	tb.logger.Error(tb.dsp.ListenWebhook(webhookURL).Error())
}

// API returns the reference to the echotron.API.
func (tb *TBB) API() echotron.API {
	return tb.api
}

// Config returns the config
func (tb *TBB) Config() *Config {
	return tb.cfg
}

// DB returns the database handle for the bot so that the database can easily be adjusted and extended.
func (tb *TBB) DB() *DB {
	return tb.db
}

// Dispatcher returns the echotron.Dispatcher.
func (tb *TBB) Dispatcher() *echotron.Dispatcher {
	return tb.dsp
}

// Server returns the http.Server.
func (tb *TBB) Server() *http.Server {
	return tb.srv
}

// SetBotCommands registers the given command list for your Telegram bot.
// Will delete registered bot commands if parameter bc is nil.
func (tb *TBB) SetBotCommands(bc []echotron.BotCommand) error {
	if bc == nil {
		_, err := tb.api.DeleteMyCommands(nil)
		return err
	}
	_, err := tb.api.SetMyCommands(nil, bc...)
	return err
}

func (tb *TBB) newBot(chatID int64, l *slog.Logger, hFn UpdateHandlerFn) *Bot {
	b := &Bot{
		app:    tb,
		chatID: chatID,
		logger: l.WithGroup("Bot"),
	}

	if b.chatID == 0 {
		panic("missing chat ID")
	}

	var err error
	b.user, err = tb.DB().FindUserByChatID(b.chatID)
	if err != nil {
		b.logger.Warn(err.Error())
		b.logger.Info(fmt.Sprintf("Creating new user with ChatID=%d", b.chatID))
		b.user = &model.User{ChatID: b.chatID, UserInfo: &model.UserInfo{}, UserPhoto: &model.UserPhoto{}}
	}

	// Create tb new UpdateHandler and set Bot reference back on handler
	b.handler = hFn()
	b.handler.SetBot(b)
	// Set the self-destruction timer
	b.dTimer = time.AfterFunc(time.Duration(tb.cfg.BotSessionTimeout)*time.Minute, b.destruct)
	b.logger.Debug(fmt.Sprintf("New Bot instance started with ChatID=%d", b.chatID))

	return b
}

func (tb *TBB) buildBot(h UpdateHandlerFn) echotron.NewBotFn {
	return func(chatId int64) echotron.Bot {
		return tb.newBot(chatId, tb.logger, h)
	}
}

func (tb *TBB) getRegistryCommand(name string) *Command {
	c, ok := tb.cmdReg[name]
	if !ok {
		return nil
	}
	return &c
}

func (tb *TBB) buildTelegramCommands() []echotron.BotCommand {
	var bc []echotron.BotCommand
	for _, c := range tb.cmdReg {
		if c.Name != "" && c.Description != "" {
			bc = append(bc, echotron.BotCommand{
				Command:     c.Name,
				Description: c.Description,
			})
		}
	}
	return bc
}
