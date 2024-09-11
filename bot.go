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
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	memberStatusJoin  = "member"
	memberStatusLeave = "kicked"

	// We only try to update user data if more than 24 hours have passed since the last update.
	updateDuration = time.Hour * 24
)

type StateFn func(*echotron.Update) StateFn

type Bot struct {
	tbot    *TBot // Backreference to TBot instance
	chatID  int64
	cmd     *Command
	handler UpdateHandler
	state   StateFn
	user    *User
	logger  *slog.Logger
	dTimer  *time.Timer // Destruction timer
	mu      sync.Mutex
}

// ChatID returns the user chatID
func (b *Bot) ChatID() int64 {
	return b.chatID
}

// Command returns the last command that was sent by the user
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
	return b.tbot.API()
}

// TBot returns the TBot reference
func (b *Bot) TBot() *TBot {
	return b.tbot
}

// DB Returns the database reference
func (b *Bot) DB() *DB {
	return b.tbot.DB()
}

// IsUserActive returns true if the user is active or false otherwise
func (b *Bot) IsUserActive() bool {
	return b.user.UserInfo.IsActive
}

// ReplaceMessage replaces the given CallbackQuery message with new Text and Keyboard
func (b *Bot) ReplaceMessage(q *echotron.CallbackQuery, text string, buttons [][]echotron.InlineKeyboardButton) {
	_, _ = b.tbot.API().EditMessageText(text, echotron.NewMessageID(b.chatID, q.Message.ID), &echotron.MessageTextOptions{ReplyMarkup: echotron.InlineKeyboardMarkup{InlineKeyboard: buttons}})
}

// DeleteMessage deletes the given CallbackQuery message
func (b *Bot) DeleteMessage(q *echotron.CallbackQuery) {
	_, _ = b.tbot.API().DeleteMessage(b.chatID, q.Message.ID)
}

// EnableUser enables the current user and updates the database
func (b *Bot) EnableUser() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.user.UserInfo.IsActive = true
	b.user.UserInfo.Status = memberStatusJoin
}

// DisableUser disable the current user and update the database
func (b *Bot) DisableUser() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.user.UserInfo.IsActive = false
	b.user.UserInfo.Status = memberStatusLeave
}

// GetUsersTimezoneOffset returns the time zone offset in seconds if the user has already provided coordinates.
// This can be used, for example, to calculate the current time of the user.
func (b *Bot) GetUsersTimezoneOffset() (int, error) {
	uInfo := b.user.UserInfo
	if uInfo.Latitude == 0 && uInfo.Longitude == 0 {
		return 0, errors.New("no user coordinates available")
	}
	return b.tbot.GetCurrentTimeOffset(uInfo.Latitude, uInfo.Longitude), nil
}

// Update is called whenever a Telegram update occurs
func (b *Bot) Update(u *echotron.Update) {
	defer b.logRecoveredPanic()

	b.resetSessionTimeout()

	// Allow only users from AllowedChatIDs to use the bot
	if len(b.tbot.cfg.AllowedChatIDs) > 0 && !slices.Contains(b.tbot.cfg.AllowedChatIDs, u.ChatID()) {
		if b.user.UserInfo.IsActive {
			b.DisableUser()
			b.DB().Save(b.user)
		}
		b.logger.Info("Access denied for user", "user", PrintAsJson(b.User(), false), "update", PrintAsJson(u, false))
		return
	}

	// Check asynchronously if we need to update user information from Telegram
	go b.updateUserData(u, updateDuration)

	// Commands always take the highest precedence
	if cmd := b.getCommand(u); cmd != nil {
		b.cmd = cmd
		if b.cmd.Handler != nil {
			b.cmd.Handler.SetBot(b)
			b.state = b.cmd.Handler.Handle()
		}
		return
	}

	// This kind of message has precedence because it disables or enables the Bot
	if u.MyChatMember != nil {
		b.state = b.handler.HandleMyChatMember(*u.MyChatMember)
		return
	}

	// If bot state is nil, we set the initial state in relation to the received update
	if b.state == nil {
		b.cmd = nil
		b.state = b.handleInitialState(u)
		return
	}

	b.state = b.state(u)
}

func (b *Bot) resetSessionTimeout() {
	st := time.Duration(b.tbot.cfg.BotSessionTimeout) * time.Minute
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

		c := b.tbot.getRegistryCommand(cmd[0])
		if c != nil && len(cmd) > 1 {
			c.Params = cmd[1:]
		}

		return c
	}
	return nil
}

// updateUser updates the user infos with the current user data from Telegram
func (b *Bot) updateUser(u *echotron.Update) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var (
		user = GetUserFromUpdate(u)
		err  error
	)

	b.user.Firstname = user.FirstName
	b.user.Lastname = user.LastName
	b.user.Username = user.Username
	b.user.LanguageCode = user.LanguageCode
	b.user.IsBot = user.IsBot
	b.user.IsPremium = user.IsPremium
	b.user.AddedToAttachmentMenu = user.AddedToAttachmentMenu
	b.user.CanJoinGroups = user.CanJoinGroups
	b.user.CanReadAllGroupMessages = user.CanReadAllGroupMessages
	b.user.SupportsInlineQueries = user.SupportsInlineQueries
	b.user.CanConnectToBusiness = user.CanConnectToBusiness
	b.user.HasMainWebApp = user.HasMainWebApp

	if b.user.UserPhoto, err = b.fetchCurrentUserPhoto(); err != nil {
		// Warn if a user photo cannot be updated but proceed anyway
		b.Log().Warn(err.Error())
	}

	return b.tbot.DB().Save(b.user).Error
}

// updateUserData updates the DB user data with data from Telegram update only if the
// chatType is "private" and more than dur time has passed since the last update.
func (b *Bot) updateUserData(u *echotron.Update, dur time.Duration) {
	// Only private user chats will be saved to the database because
	// we don't want to save channel or group infos as users in the database.
	if GetChatTypeFromUpdate(u) != ChatTypePrivate {
		return
	}

	// We only update user data in the database if more than dur seconds have elapsed.
	if time.Since(b.user.UpdatedAt) < dur {
		return
	}

	if err := b.updateUser(u); err != nil {
		b.Log().Error(err.Error())
	}
}

func (b *Bot) destruct() {
	b.tbot.dsp.DelSession(b.chatID)
	b.logger.Info(fmt.Sprintf("Deleted bot instance with ChatID=%d", b.chatID))
}

// fetchCurrentUserPhoto tries to update the current users photo with the data from Telegram
func (b *Bot) fetchCurrentUserPhoto() (*UserPhoto, error) {
	userPhoto := &UserPhoto{}

	res, err := b.tbot.api.GetUserProfilePhotos(b.user.ChatID, &echotron.UserProfileOptions{Offset: 0, Limit: 1})
	if err != nil {
		return userPhoto, err
	}

	if !res.Ok {
		return userPhoto, errors.New("could not get user profile")
	}

	b.logger.Debug("GetUserProfilePhotos request successful!", "totalPhotos", res.Result.TotalCount)

	if len(res.Result.Photos) == 0 {
		return userPhoto, nil
	}

	newestPhotoSizes := res.Result.Photos[0]
	biggestPhotoSize := newestPhotoSizes[len(newestPhotoSizes)-1]

	fileID, err := b.tbot.api.GetFile(biggestPhotoSize.FileID)
	if err != nil {
		return userPhoto, err
	}

	photoURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.tbot.cfg.Telegram.BotToken, fileID.Result.FilePath)
	b.logger.Debug(fileID.Result.FilePath)
	fileRes, err := http.Get(photoURL)
	if err != nil {
		return userPhoto, err
	}

	data, err := io.ReadAll(fileRes.Body)
	if err != nil {
		return userPhoto, err
	}

	b.logger.Info("Updated user photo", "userID", b.user.ID)

	userPhoto = &UserPhoto{
		UserID:       b.user.ID,
		FileID:       biggestPhotoSize.FileID,
		FileUniqueID: biggestPhotoSize.FileUniqueID,
		FileSize:     biggestPhotoSize.FileSize,
		FileHash:     fmt.Sprintf("%x", md5.Sum(data)),
		FileData:     data,
		Width:        biggestPhotoSize.Width,
		Height:       biggestPhotoSize.Height,
	}

	return userPhoto, nil
}
