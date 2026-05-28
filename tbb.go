// Package tbb provides a base for creating custom Telegram bots.
// It is database-agnostic and uses PaulSonOfLars/gotgbot/v2.
package tbb

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	timezone "github.com/evanoberholster/timezoneLookup/v2"
	"github.com/gabriel-vasile/mimetype"
)

type TBot struct {
	store      UserStore
	ctx        context.Context
	cfg        *Config
	logger     *slog.Logger
	cmdReg     CommandRegistry
	hFn        UpdateHandlerFn
	bot        *gotgbot.Bot
	updater    *ext.Updater
	tzc        *timezone.Timezonecache
	srv        *http.Server
	sessions   map[int64]*Bot
	sessionsMu sync.RWMutex
}

type Option func(*TBot)

// New creates a new Telegram bot based on the given configuration.
// It uses functional options for configuration.
func New(opts ...Option) *TBot {
	tbot := &TBot{
		ctx:      context.Background(),
		cmdReg:   CommandRegistry{},
		hFn:      func() UpdateHandler { return &DefaultUpdateHandler{} },
		logger:   nil,
		tzc:      loadTimezoneCache(),
		sessions: make(map[int64]*Bot),
	}

	// Loop through each option
	for _, opt := range opts {
		opt(tbot)
	}

	if tbot.cfg == nil {
		tbot.cfg = &Config{}
	}

	if tbot.logger == nil {
		tbot.logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource: true,
			Level:     getLogLevel(tbot.cfg.LogLevel),
		}))
	}

	if tbot.store == nil {
		tbot.store = NewInMemoryStore()
	}

	if tbot.cfg.Telegram.BotToken == "" {
		panic("tbot config is missing telegram bot token")
	}

	var err error
	tbot.bot, err = gotgbot.NewBot(tbot.cfg.Telegram.BotToken, &gotgbot.BotOpts{
		DisableTokenCheck: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create telegram bot: %v", err))
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

// WithToken sets the Telegram bot token configuration directly.
func WithToken(token string) Option {
	return func(app *TBot) {
		if app.cfg == nil {
			app.cfg = &Config{}
		}
		app.cfg.Telegram.BotToken = token
	}
}

// WithAllowedChatIDs sets the AllowedChatIDs directly.
func WithAllowedChatIDs(ids []int64) Option {
	return func(app *TBot) {
		if app.cfg == nil {
			app.cfg = &Config{}
		}
		app.cfg.AllowedChatIDs = ids
	}
}

// WithSessionTimeout sets the bot session timeout in minutes.
func WithSessionTimeout(minutes int) Option {
	return func(app *TBot) {
		if app.cfg == nil {
			app.cfg = &Config{}
		}
		app.cfg.BotSessionTimeout = minutes
	}
}

// WithLogLevel sets the logging level.
func WithLogLevel(level string) Option {
	return func(app *TBot) {
		if app.cfg == nil {
			app.cfg = &Config{}
		}
		app.cfg.LogLevel = level
	}
}

// WithUserStore configures a custom user store implementation.
func WithUserStore(store UserStore) Option {
	return func(app *TBot) {
		app.store = store
	}
}

// WithHandlerFunc option can be used to override the default UpdateHandlerFn for custom gotgbot.Update message handling.
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
	if err := tb.SetBotCommands(tb.buildTelegramCommands()); err != nil {
		tb.logger.Error("Cannot set bot commands!")
		panic(err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			tb.logger.Error("Dispatcher error", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: -1,
	})
	dispatcher.AddHandler(sessionDispatcherHandler{tbot: tb})

	tb.updater = ext.NewUpdater(dispatcher, nil)

	tb.logger.Info("Start polling updates")
	err := tb.updater.StartPolling(tb.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
	})
	if err != nil {
		panic(err)
	}

	// Handle OS signals for graceful shutdown
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	tb.logger.Info("Stopping polling updates...")
	if err := tb.updater.Stop(); err != nil {
		tb.logger.Error("Failed to stop updater", "error", err)
	}
	tb.logger.Info("Bot stopped")
}

// StartWithWebhook starts the Telegram bot server with a given webhook url path and listen address.
func (tb *TBot) StartWithWebhook(webhookPath string, listenAddr string) {
	if err := tb.SetBotCommands(tb.buildTelegramCommands()); err != nil {
		tb.logger.Error("Cannot set bot commands!")
		panic(err)
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			tb.logger.Error("Dispatcher error", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: -1,
	})
	dispatcher.AddHandler(sessionDispatcherHandler{tbot: tb})

	tb.updater = ext.NewUpdater(dispatcher, nil)

	tb.logger.Info("Starting webhook server", "addr", listenAddr, "path", webhookPath)
	err := tb.updater.StartWebhook(tb.bot, webhookPath, ext.WebhookOpts{
		ListenAddr: listenAddr,
	})
	if err != nil {
		panic(err)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

	tb.logger.Info("Stopping webhook...")
	if err := tb.updater.Stop(); err != nil {
		tb.logger.Error("Failed to stop updater", "error", err)
	}
	tb.logger.Info("Bot stopped")
}

func (tb *TBot) handleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	chatID := getChatID(ctx.Update)
	if chatID == 0 {
		tb.logger.Warn("update has no chat ID", "updateID", ctx.UpdateId)
		return nil
	}

	tb.sessionsMu.Lock()
	session, exists := tb.sessions[chatID]
	if !exists {
		session = tb.newBot(chatID, tb.logger, tb.hFn)
		tb.sessions[chatID] = session
	}
	tb.sessionsMu.Unlock()

	session.Update(ctx.Update)
	return nil
}

// API returns the reference to the gotgbot.Bot API wrapper.
func (tb *TBot) API() *gotgbot.Bot {
	return tb.bot
}

// Config returns the config
func (tb *TBot) Config() *Config {
	return tb.cfg
}

// Store returns the user store.
func (tb *TBot) Store() UserStore {
	return tb.store
}

// Updater returns the gotgbot updater.
func (tb *TBot) Updater() *ext.Updater {
	return tb.updater
}

// DownloadFile downloads a file from Telegram by a given fileID
func (tb *TBot) DownloadFile(fileID string) (*File, error) {
	fileIDRes, err := tb.bot.GetFile(fileID, nil)
	if err != nil {
		return nil, err
	}
	tb.logger.Debug("GetFile request successful", "filePath", fileIDRes.FilePath)

	photoURL := fileIDRes.URL(tb.bot, nil)
	fileRes, err := http.Get(photoURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = fileRes.Body.Close() }()

	fileData, err := io.ReadAll(fileRes.Body)
	if err != nil {
		return nil, err
	}

	mime := mimetype.Detect(fileData)
	f := &File{
		UniqueID:  fileIDRes.FileUniqueId,
		Extension: mime.Extension(),
		MimeType:  mime.String(),
		Hash:      fmt.Sprintf("%x", md5.Sum(fileData)),
		Size:      int64(fileIDRes.FileSize),
		Data:      fileData,
	}

	return f, nil
}

// SetBotCommands registers the given command list for your Telegram bot.
// Will delete registered bot commands if parameter bc is nil.
func (tb *TBot) SetBotCommands(bc []gotgbot.BotCommand) error {
	if bc == nil {
		_, err := tb.bot.DeleteMyCommands(nil)
		return err
	}
	_, err := tb.bot.SetMyCommands(bc, nil)
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
	b.user, err = tb.Store().FindUserByChatID(b.chatID)
	if err != nil {
		b.logger.Warn(err.Error())
		b.logger.Info(fmt.Sprintf("Creating new user with ChatID=%d", b.chatID))
		b.user = &User{ChatID: b.chatID, UserInfo: &UserInfo{}, UserPhoto: &UserPhoto{}}
	}

	// Create new UpdateHandler and set Bot reference back on handler
	b.handler = hFn()
	b.handler.SetBot(b)
	// Set the self-destruction timer
	b.dTimer = time.AfterFunc(time.Duration(tb.cfg.BotSessionTimeout)*time.Minute, b.destruct)
	b.logger.Debug(fmt.Sprintf("New Bot instance started with ChatID=%d", b.chatID))

	return b
}

func (tb *TBot) getRegistryCommand(name string) *Command {
	c, ok := tb.cmdReg[name]
	if !ok {
		return nil
	}
	return &c
}

func (tb *TBot) buildTelegramCommands() []gotgbot.BotCommand {
	var bc []gotgbot.BotCommand
	for _, c := range tb.cmdReg {
		if c.Name != "" && c.Description != "" {
			// Strip leading slash for Telegram command settings
			name := strings.TrimPrefix(c.Name, "/")
			bc = append(bc, gotgbot.BotCommand{
				Command:     name,
				Description: c.Description,
			})
		}
	}
	return bc
}

type sessionDispatcherHandler struct {
	tbot *TBot
}

func (h sessionDispatcherHandler) CheckUpdate(b *gotgbot.Bot, ctx *ext.Context) bool {
	return true
}

func (h sessionDispatcherHandler) HandleUpdate(b *gotgbot.Bot, ctx *ext.Context) error {
	return h.tbot.handleUpdate(b, ctx)
}

func (h sessionDispatcherHandler) Name() string {
	return "tbb.SessionDispatcherHandler"
}
