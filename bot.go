package tbb

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/NicoNex/echotron/v3"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

const (
	memberStatusJoin  = "member"
	memberStatusLeave = "kicked"
)

type StateFn func(*echotron.Update) StateFn

type Bot struct {
	app     *App
	chatID  int64
	cmd     *Command
	handler UpdateHandler
	state   StateFn
	user    *User
	logger  *slog.Logger
	dTimer  *time.Timer // Destruction timer
}

type BotOption func(*Bot)

// ChatID returns the user chatID
func (b *Bot) ChatID() int64 {
	return b.chatID
}

func (b *Bot) Command() *Command {
	return b.cmd
}

// User returns the User associated with the bot
func (b *Bot) User() *User {
	return b.user
}

// Log can be used to access the logger.
func (b *Bot) Log() *slog.Logger {
	return b.logger
}

// API returns the echotron.API
func (b *Bot) API() echotron.API {
	return b.app.API()
}

// App returns the App reference
func (b *Bot) App() *App {
	return b.app
}

// DB Returns the database reference
func (b *Bot) DB() *DB {
	return b.app.DB()
}

func (b *Bot) IsUserActive() bool {
	return b.user.UserInfo.IsActive
}

// ReplaceMessage replaces the given CallbackQuery message with new Text and Keyboard
func (b *Bot) ReplaceMessage(q *echotron.CallbackQuery, text string, buttons [][]echotron.InlineKeyboardButton) {
	_, _ = b.app.API().EditMessageText(text, echotron.NewMessageID(b.chatID, q.Message.ID), &echotron.MessageTextOptions{ReplyMarkup: echotron.InlineKeyboardMarkup{InlineKeyboard: buttons}})
}

// DeleteMessage deletes the given CallbackQuery message
func (b *Bot) DeleteMessage(q *echotron.CallbackQuery) {
	_, _ = b.app.API().DeleteMessage(b.chatID, q.Message.ID)
}

// EnableUser enables current user and updates the database
func (b *Bot) EnableUser() {
	b.user.UserInfo.IsActive = true
	b.user.UserInfo.Status = memberStatusJoin
	b.app.DB().Save(b.user)
}

// DisableUser disable current user and updates the database
func (b *Bot) DisableUser() {
	b.user.UserInfo.IsActive = false
	b.user.UserInfo.Status = memberStatusLeave
	b.app.DB().Save(b.user)
}

// GetUsernameFromMessage returns the Firstname or Username from the telegram user of the given message
func (b *Bot) GetUsernameFromMessage(m *echotron.Message) string {
	switch {
	case m == nil:
		return ""
	case m.From.FirstName != "":
		return m.From.FirstName
	case m.From.Username != "":
		return m.From.Username
	default:
		return ""
	}
}

// GetUsersTimezoneOffset returns the time zone offset in seconds if the user has already provided coordinates.
// This can be used, for example, to calculate the current time of the user.
func (b *Bot) GetUsersTimezoneOffset() (int, error) {
	uInfo := b.user.UserInfo
	if uInfo.Latitude == 0 && uInfo.Longitude == 0 {
		return 0, errors.New("no user coordinates available")
	}
	return b.app.GetCurrentTimeOffset(uInfo.Latitude, uInfo.Longitude), nil
}

// Update is called whenever a Telegram update occurs
func (b *Bot) Update(u *echotron.Update) {
	defer b.logRecoveredPanic()

	b.resetSessionTimeout()

	// Check if we need to update user information from telegram
	if time.Since(b.user.UpdatedAt).Hours() > 24*7 {
		go func() {
			b.updateUser(u)
			b.updateUserPhoto()
			b.app.DB().Save(b.user)
		}()
	}

	// Commands always take the highest precedence
	if cmd := b.getCommand(u); cmd != nil {
		b.cmd = cmd
		if b.cmd.Handler != nil {
			b.state = b.cmd.Handler.Handle(b)
		}
		return
	}

	// This kind of message has precedence because it disables or enables the Bot
	if u.MyChatMember != nil {
		b.state = b.handler.HandleMyChatMember(*u.MyChatMember)
		return
	}

	// If state is nil, we set the initial state in relation to the received update
	if b.state == nil {
		b.cmd = nil
		b.state = b.handleInitialState(u)
		return
	}

	b.state = b.state(u)
}

func (b *Bot) resetSessionTimeout() {
	st := time.Duration(b.App().cfg.BotSessionTimeout) * time.Minute
	b.dTimer.Reset(st)
	b.logger.Debug(fmt.Sprintf("Bot istance with ChatID=%d will exire at %s", b.chatID, time.Now().Add(st)))
}

func (b *Bot) logRecoveredPanic() {
	if r := recover(); r != nil {
		b.logger.Error("Recovered panic in update", "panic", r, "stack", string(debug.Stack()))
	}
}

func (b *Bot) handleInitialState(u *echotron.Update) StateFn {
	switch {
	case u.Message != nil:
		return b.handler.HandleMessage(*u.Message)
	case u.EditedMessage != nil:
		return b.handler.HandleEditedMessage(*u.EditedMessage)
	case u.ChannelPost != nil:
		return b.handler.HandleChannelPost(*u.ChannelPost)
	case u.EditedChannelPost != nil:
		return b.handler.HandleEditedChannelPost(*u.EditedChannelPost)
	case u.InlineQuery != nil:
		return b.handler.HandleInlineQuery(*u.InlineQuery)
	case u.ChosenInlineResult != nil:
		return b.handler.HandleChosenInlineResult(*u.ChosenInlineResult)
	case u.CallbackQuery != nil:
		return b.handler.HandleCallbackQuery(*u.CallbackQuery)
	case u.ShippingQuery != nil:
		return b.handler.HandleShippingQuery(*u.ShippingQuery)
	case u.PreCheckoutQuery != nil:
		return b.handler.HandlePreCheckoutQuery(*u.PreCheckoutQuery)
	case u.ChatMember != nil:
		return b.handler.HandleChatMember(*u.ChatMember)
	case u.ChatJoinRequest != nil:
		return b.handler.HandleChatJoinRequest(*u.ChatJoinRequest)
	default:
		return b.handleUnknown(u)
	}
}

func (b *Bot) handleUnknown(u *echotron.Update) StateFn {
	jsonStr, err := json.Marshal(u)
	if err != nil {
		b.logger.Error(err.Error())
	}
	b.logger.Error("update has an unknown type", "update", string(jsonStr))
	return nil
}

// Returns the command slice if available or nil if no command exists
func (b *Bot) getCommand(u *echotron.Update) *Command {
	var text string
	switch {
	case u.Message != nil:
		text = u.Message.Text
	case u.EditedMessage != nil:
		text = u.EditedMessage.Text
	}
	text = strings.TrimSpace(text)
	if strings.Index(text, "/") == 0 {
		re := regexp.MustCompile(`\s{2,}`)
		cmd := strings.Split(re.ReplaceAllString(text, " "), " ")

		c := b.app.getRegistryCommand(cmd[0])
		if c != nil && len(cmd) > 1 {
			c.Params = cmd[1:]
		}

		return c
	}
	return nil
}

func (b *Bot) updateUser(u *echotron.Update) {
	var user = GetUserFromUpdate(u)
	b.user.Firstname = user.FirstName
	b.user.Lastname = user.LastName
	b.user.Username = user.Username
	b.user.IsBot = user.IsBot
	b.user.IsPremium = user.IsPremium
	b.user.LanguageCode = user.LanguageCode
}

func (b *Bot) destruct() {
	b.App().dsp.DelSession(b.chatID)
	b.logger.Info(fmt.Sprintf("Deleted bot instance with ChatID=%d", b.chatID))
}

func (b *Bot) updateUserPhoto() {
	res, err := b.app.api.GetUserProfilePhotos(b.user.ChatID, &echotron.UserProfileOptions{Offset: 0, Limit: 1})
	if err != nil {
		b.logger.Error(err.Error())
		return
	}

	if res.Ok {
		b.logger.Debug("GetUserProfilePhotos request successful!", "totalPhotos", res.Result.TotalCount)

		if len(res.Result.Photos) == 0 {
			return
		}

		newestPhotoSizes := res.Result.Photos[0]
		biggestPhotoSize := newestPhotoSizes[len(newestPhotoSizes)-1]

		fileID, err := b.app.api.GetFile(biggestPhotoSize.FileID)
		if err != nil {
			b.logger.Error(err.Error())
		}

		photoURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.app.cfg.Telegram.BotToken, fileID.Result.FilePath)
		b.logger.Debug(fileID.Result.FilePath)
		fileRes, err := http.Get(photoURL)
		if err != nil {
			b.logger.Error(err.Error())
		}

		data, err := io.ReadAll(fileRes.Body)
		if err != nil {
			b.logger.Error(err.Error())
		}

		b.user.UserPhoto = &UserPhoto{
			UserID:       b.user.ID,
			FileID:       biggestPhotoSize.FileID,
			FileUniqueID: biggestPhotoSize.FileUniqueID,
			FileSize:     biggestPhotoSize.FileSize,
			FileHash:     fmt.Sprintf("%x", md5.Sum(data)),
			FileData:     data,
			Width:        biggestPhotoSize.Width,
			Height:       biggestPhotoSize.Height,
		}

		b.logger.Info("Updated user photo", "user", b.user)
	} else {
		b.logger.Debug("GetUserProfilePhotos failed!")
	}
}

type InlineKeyboardButton struct {
	Text string `json:"text"`
	Data string `json:"data"`
}
