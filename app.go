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
	"log/slog"
	"os"
	"time"
)

type App struct {
	db     *DB
	dsp    *echotron.Dispatcher
	ctx    context.Context
	cfg    *Config
	logger *slog.Logger
	cmdReg CommandRegistry
	hFn    UpdateHandlerFn
	api    echotron.API // Telegram api
	tzc    *timezone.Timezonecache
}

type AppOption func(*App)

// NewApp creates a new Telegram bot based on the given configuration.
// It uses functional options for configuration.
func NewApp(opts ...AppOption) *App {
	app := &App{
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

	app.api = echotron.NewAPI(app.cfg.Telegram.BotToken)
	app.db = newDB(app.cfg, app.logger)
	app.dsp = echotron.NewDispatcher(app.cfg.Telegram.BotToken, app.buildBot(app.hFn))

	return app
}

// WithConfig is the only required option, because it provides the config for the app to function properly.
func WithConfig(cfg *Config) AppOption {
	return func(app *App) {
		app.cfg = cfg
	}
}

// WithCommands is used for providing and registering custom bot commands.
// Bot commands always start with a / like /start and a Handler, which implements the CommandHandler interface.
// If you want a command to be available in the command list on Telegram,
// the provided Command must contain a Description.
func WithCommands(commands []Command) AppOption {
	return func(app *App) {
		app.cmdReg = buildCommandRegistry(commands)
	}
}

// WithHandlerFunc option can be used to override the default UpdateHandlerFn for custom echotron.Update message handling.
func WithHandlerFunc(hFn UpdateHandlerFn) AppOption {
	return func(app *App) {
		app.hFn = hFn
	}
}

// WithLogger option can be used to override the default logger with a custom one.
func WithLogger(l *slog.Logger) AppOption {
	return func(app *App) {
		app.logger = l
	}
}

// Start starts the Telegram bot server in poll mode
func (a *App) Start() {
	if err := a.SetBotCommands(a.buildTelegramCommands()); err != nil {
		a.logger.Error("Cannot set bot commands!")
		os.Exit(1)
	}
	a.logger.Info("Start dispatcher...")
	a.logger.Error(a.dsp.Poll().Error())
}

// StartWithWebhook starts the Telegram bot server with a given webhook url.
func (a *App) StartWithWebhook(webhookURL string) {
	if err := a.SetBotCommands(a.buildTelegramCommands()); err != nil {
		a.logger.Error("Cannot set bot commands!")
		os.Exit(1)
	}
	a.logger.Info("Start dispatcher...")

	if webhookURL == "" {
		panic("webhook url is empty")
	}

	a.logger.Error(a.dsp.ListenWebhook(webhookURL).Error())
}

// API returns the reference to the echotron.API.
func (a *App) API() echotron.API {
	return a.api
}

// Config returns the config
func (a *App) Config() *Config {
	return a.cfg
}

// DB returns the database handle for the bot so that the database can easily be adjusted and extended.
func (a *App) DB() *DB {
	return a.db
}

// SetBotCommands registers the given command list for your Telegram bot.
// Will delete registered bot commands if parameter bc is nil.
func (a *App) SetBotCommands(bc []echotron.BotCommand) error {
	if bc == nil {
		_, err := a.api.DeleteMyCommands(nil)
		return err
	}
	_, err := a.api.SetMyCommands(nil, bc...)
	return err
}

func (a *App) newBot(chatID int64, l *slog.Logger, hFn UpdateHandlerFn) *Bot {
	b := &Bot{
		app:    a,
		chatID: chatID,
		logger: l.WithGroup("Bot"),
	}

	if b.chatID == 0 {
		panic("missing chat ID")
	}

	var err error
	b.user, err = a.DB().FindUserByChatID(b.chatID)
	if err != nil {
		b.logger.Warn(err.Error())
		b.logger.Info(fmt.Sprintf("Creating new user with ChatID=%d", b.chatID))
		b.user = &model.User{ChatID: b.chatID, UserInfo: &model.UserInfo{}, UserPhoto: &model.UserPhoto{}}
	}

	// Create a new UpdateHandler and set Bot reference back on handler
	b.handler = hFn()
	b.handler.SetBot(b)
	// Set the self-destruction timer
	b.dTimer = time.AfterFunc(time.Duration(a.cfg.BotSessionTimeout)*time.Minute, b.destruct)
	b.logger.Debug(fmt.Sprintf("New Bot instance started with ChatID=%d", b.chatID))

	return b
}

func (a *App) buildBot(h UpdateHandlerFn) echotron.NewBotFn {
	return func(chatId int64) echotron.Bot {
		return a.newBot(chatId, a.logger, h)
	}
}

func (a *App) getRegistryCommand(name string) *Command {
	c, ok := a.cmdReg[name]
	if !ok {
		return nil
	}
	return &c
}

func (a *App) buildTelegramCommands() []echotron.BotCommand {
	var bc []echotron.BotCommand
	for _, c := range a.cmdReg {
		if c.Name != "" && c.Description != "" {
			bc = append(bc, echotron.BotCommand{
				Command:     c.Name,
				Description: c.Description,
			})
		}
	}
	return bc
}
