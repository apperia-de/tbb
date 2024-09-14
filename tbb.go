// Package tbb provides a base for creating custom Telegram bots.
// It can be used to spin up a custom bot in one minute which is capable of
// handling bot users via an implemented sqlite database,
// but can easily be switched to use mysql or postgres instead.
package tbb

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/NicoNex/echotron/v3"
	timezone "github.com/evanoberholster/timezoneLookup/v2"
	"github.com/gabriel-vasile/mimetype"
	"gorm.io/gorm"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type TBot struct {
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

type Option func(*TBot)

// New creates a new Telegram bot based on the given configuration.
// It uses functional options for configuration.
func New(opts ...Option) *TBot {
	tbot := &TBot{
		ctx:    context.Background(),
		cmdReg: CommandRegistry{},
		hFn:    func() UpdateHandler { return &DefaultUpdateHandler{} },
		logger: nil,
		tzc:    loadTimezoneCache(),
	}

	// Loop through each option
	for _, opt := range opts {
		opt(tbot)
	}

	if tbot.cfg == nil {
		panic("tbot config is missing")
	}

	if tbot.logger == nil {
		tbot.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     getLogLevel(tbot.cfg.LogLevel),
		}))
	}

	tbot.db = NewDB(tbot.cfg, &gorm.Config{FullSaveAssociations: true})
	tbot.api = echotron.NewAPI(tbot.cfg.Telegram.BotToken)
	tbot.dsp = echotron.NewDispatcher(tbot.cfg.Telegram.BotToken, tbot.buildBot(tbot.hFn))
	if tbot.srv != nil {
		tbot.dsp.SetHTTPServer(tbot.srv)
	}

	// Initialize database tables
	if err := tbot.db.AutoMigrate(&User{}, &UserInfo{}, &UserPhoto{}); err != nil {
		panic(err)
	}

	return tbot
}

// WithConfig is the only required option because it provides the config for the tbot to function properly.
func WithConfig(cfg *Config) Option {
	return func(app *TBot) {
		app.cfg = cfg
	}
}

// WithCommands is used for providing and registering custom bot commands.
// Bot commands always start with a / like /start and a Handler, which implements the CommandHandler interface.
// If you want a command to be available in the command list on Telegram,
// the provided Command must contain a Description.
func WithCommands(commands []Command) Option {
	return func(app *TBot) {
		app.cmdReg = buildCommandRegistry(commands)
	}
}

// WithHandlerFunc option can be used to override the default UpdateHandlerFn for custom echotron.Update message handling.
func WithHandlerFunc(hFn UpdateHandlerFn) Option {
	return func(app *TBot) {
		app.hFn = hFn
	}
}

// WithLogger option can be used to override the default logger with a custom one.
func WithLogger(l *slog.Logger) Option {
	return func(app *TBot) {
		app.logger = l
	}
}

// WithServer option can be used add a custom http.Server to the dispatcher
func WithServer(s *http.Server) Option {
	return func(app *TBot) {
		app.srv = s
	}
}

// Start starts the Telegram bot server in poll mode
func (tb *TBot) Start() {
	var err error
	if err = tb.SetBotCommands(tb.buildTelegramCommands()); err != nil {
		tb.logger.Error("Cannot set bot commands!")
		panic(err)
	}

	if tb.srv == nil {
		tb.logger.Info("Start dispatcher")
		tb.logger.Error(tb.dsp.Poll().Error())
		return
	}

	// If we have a custom web server, we run the polling in a separate go routine.
	go func() {
		tb.logger.Info("Start dispatcher")
		tb.logger.Error(tb.dsp.Poll().Error())
	}()

	go shutdownServerOnSignal(tb.srv)

	tb.logger.Info("Start server")
	err = tb.srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		tb.logger.Error(err.Error())
		return
	}
	tb.logger.Info("Server closed")
}

// StartWithWebhook starts the Telegram bot server with a given webhook url.
func (tb *TBot) StartWithWebhook(webhookURL string) {
	var err error
	if err = tb.SetBotCommands(tb.buildTelegramCommands()); err != nil {
		tb.logger.Error("Cannot set bot commands!")
		panic(err)
	}
	if webhookURL == "" {
		panic("webhook url is empty")
	}

	tb.logger.Info(fmt.Sprintf("Start dispatcher and server with webhook: %q", webhookURL))

	if tb.srv != nil {
		go shutdownServerOnSignal(tb.srv)
	}

	err = tb.dsp.ListenWebhook(webhookURL)
	if !errors.Is(err, http.ErrServerClosed) {
		tb.logger.Error(err.Error())
		return
	}
	tb.logger.Info("Server closed")
}

// API returns the reference to the echotron.API.
func (tb *TBot) API() echotron.API {
	return tb.api
}

// Config returns the config
func (tb *TBot) Config() *Config {
	return tb.cfg
}

// DB returns the database handle for the bot so that the database can easily be adjusted and extended.
func (tb *TBot) DB() *DB {
	return tb.db
}

// Dispatcher returns the echotron.Dispatcher.
func (tb *TBot) Dispatcher() *echotron.Dispatcher {
	return tb.dsp
}

// Server returns the http.Server.
func (tb *TBot) Server() *http.Server {
	return tb.srv
}

// DownloadFile downloads a file from Telegram by a given fileID
func (tb *TBot) DownloadFile(fileID string) (*File, error) {
	fileIDRes, err := tb.API().GetFile(fileID)
	if err != nil {
		return nil, err
	}
	tb.logger.Debug(fileIDRes.Description)

	fileData, err := tb.API().DownloadFile(fileIDRes.Result.FilePath)
	if err != nil {
		return nil, err
	}

	mime := mimetype.Detect(fileData)
	f := &File{
		UniqueID:  fileIDRes.Result.FileUniqueID,
		Extension: mime.Extension(),
		MimeType:  mime.String(),
		Hash:      fmt.Sprintf("%x", md5.Sum(fileData)),
		Size:      fileIDRes.Result.FileSize,
		Data:      fileData,
	}

	return f, nil
}

// SetBotCommands registers the given command list for your Telegram bot.
// Will delete registered bot commands if parameter bc is nil.
func (tb *TBot) SetBotCommands(bc []echotron.BotCommand) error {
	if bc == nil {
		_, err := tb.api.DeleteMyCommands(nil)
		return err
	}
	_, err := tb.api.SetMyCommands(nil, bc...)
	return err
}

func (tb *TBot) newBot(chatID int64, l *slog.Logger, hFn UpdateHandlerFn) *Bot {
	b := &Bot{
		tbot:   tb,
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
		b.user = &User{ChatID: b.chatID, UserInfo: &UserInfo{}, UserPhoto: &UserPhoto{}}
	}

	// Create tb new UpdateHandler and set Bot reference back on handler
	b.handler = hFn()
	b.handler.SetBot(b)
	// Set the self-destruction timer
	b.dTimer = time.AfterFunc(time.Duration(tb.cfg.BotSessionTimeout)*time.Minute, b.destruct)
	b.logger.Debug(fmt.Sprintf("New Bot instance started with ChatID=%d", b.chatID))

	return b
}

func (tb *TBot) buildBot(h UpdateHandlerFn) echotron.NewBotFn {
	return func(chatId int64) echotron.Bot {
		return tb.newBot(chatId, tb.logger, h)
	}
}

func (tb *TBot) getRegistryCommand(name string) *Command {
	c, ok := tb.cmdReg[name]
	if !ok {
		return nil
	}
	return &c
}

func (tb *TBot) buildTelegramCommands() []echotron.BotCommand {
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

// shutdownServerOnSignal gracefully shuts down server on SIGINT or SIGTERM
func shutdownServerOnSignal(srv *http.Server) {
	termChan := make(chan os.Signal, 1) // Channel for terminating the tbot via os.Interrupt signal
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan
	// Perform some cleanup...
	if err := srv.Shutdown(context.Background()); err != nil {
		panic(err)
	}
}
