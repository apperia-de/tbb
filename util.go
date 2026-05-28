package tbb

import (
	"encoding/json"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"log/slog"
	"strings"
)

const (
	ChatTypePrivate    ChatType = "private"
	ChatTypeChannel    ChatType = "channel"
	ChatTypeGroup      ChatType = "group"
	ChatTypeSuperGroup ChatType = "supergroup"
	ChatTypeUnknown    ChatType = "unknown"
)

type ChatType string

type InlineKeyboardButton struct {
	Text string `json:"text"`
	Data string `json:"data"`
}

// getChatID extracts the unique chat ID from a given gotgbot.Update
func getChatID(u *gotgbot.Update) int64 {
	switch {
	case u.Message != nil:
		return u.Message.Chat.Id
	case u.EditedMessage != nil:
		return u.EditedMessage.Chat.Id
	case u.ChannelPost != nil:
		return u.ChannelPost.Chat.Id
	case u.EditedChannelPost != nil:
		return u.EditedChannelPost.Chat.Id
	case u.InlineQuery != nil:
		return u.InlineQuery.From.Id
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From.Id
	case u.CallbackQuery != nil:
		if u.CallbackQuery.Message != nil {
			return u.CallbackQuery.Message.GetChat().Id
		}
		return u.CallbackQuery.From.Id
	case u.ShippingQuery != nil:
		return u.ShippingQuery.From.Id
	case u.PreCheckoutQuery != nil:
		return u.PreCheckoutQuery.From.Id
	case u.PollAnswer != nil:
		return u.PollAnswer.User.Id
	case u.MyChatMember != nil:
		return u.MyChatMember.Chat.Id
	case u.ChatMember != nil:
		return u.ChatMember.Chat.Id
	case u.ChatJoinRequest != nil:
		return u.ChatJoinRequest.Chat.Id
	default:
		return 0
	}
}

// GetUserFromUpdate returns the gotgbot.User from a given gotgbot.Update
func GetUserFromUpdate(u *gotgbot.Update) gotgbot.User {
	switch {
	case u.Message != nil:
		if u.Message.From != nil {
			return *u.Message.From
		}
	case u.EditedMessage != nil:
		if u.EditedMessage.From != nil {
			return *u.EditedMessage.From
		}
	case u.ChannelPost != nil:
		if u.ChannelPost.From != nil {
			return *u.ChannelPost.From
		}
	case u.EditedChannelPost != nil:
		if u.EditedChannelPost.From != nil {
			return *u.EditedChannelPost.From
		}
	case u.InlineQuery != nil:
		return u.InlineQuery.From
	case u.ChosenInlineResult != nil:
		return u.ChosenInlineResult.From
	case u.CallbackQuery != nil:
		return u.CallbackQuery.From
	case u.ShippingQuery != nil:
		return u.ShippingQuery.From
	case u.PreCheckoutQuery != nil:
		return u.PreCheckoutQuery.From
	case u.MyChatMember != nil:
		return u.MyChatMember.From
	case u.ChatMember != nil:
		return u.ChatMember.From
	case u.ChatJoinRequest != nil:
		return u.ChatJoinRequest.From
	}
	return gotgbot.User{Id: getChatID(u)}
}

// GetChatTypeFromUpdate returns the ChatType from a given gotgbot.Update
func GetChatTypeFromUpdate(u *gotgbot.Update) ChatType {
	convertToChatType := func(input string) ChatType {
		switch ChatType(input) {
		case ChatTypeChannel, ChatTypeGroup, ChatTypeSuperGroup, ChatTypePrivate:
			return ChatType(input)
		default:
			return ChatTypeUnknown
		}
	}

	var ct = ChatTypeUnknown
	switch {
	case u.Message != nil:
		ct = convertToChatType(u.Message.Chat.Type)
	case u.EditedMessage != nil:
		ct = convertToChatType(u.EditedMessage.Chat.Type)
	case u.ChannelPost != nil:
		ct = convertToChatType(u.ChannelPost.Chat.Type)
	case u.EditedChannelPost != nil:
		ct = convertToChatType(u.EditedChannelPost.Chat.Type)
	case u.InlineQuery != nil:
		ct = convertToChatType(u.InlineQuery.ChatType)
	case u.MyChatMember != nil:
		ct = convertToChatType(u.MyChatMember.Chat.Type)
	case u.ChatMember != nil:
		ct = convertToChatType(u.ChatMember.Chat.Type)
	case u.ChatJoinRequest != nil:
		ct = convertToChatType(u.ChatJoinRequest.Chat.Type)
	}
	return ct
}

// BuildInlineKeyboardButtonRow helper function for creating Telegram inline keyboards
func BuildInlineKeyboardButtonRow(buttons []InlineKeyboardButton) []gotgbot.InlineKeyboardButton {
	var res []gotgbot.InlineKeyboardButton
	for _, b := range buttons {
		res = append(res, gotgbot.InlineKeyboardButton{Text: b.Text, CallbackData: b.Data})
	}
	return res
}

// PrintAsJson returns the JSON representation of a given go struct.
func PrintAsJson(v any, indent bool) string {
	var jsonStr []byte
	if indent {
		jsonStr, _ = json.MarshalIndent(v, "", "  ")
	} else {
		jsonStr, _ = json.Marshal(v)
	}

	return string(jsonStr)
}

func buildCommandRegistry(commands []Command) CommandRegistry {
	cmdReg := CommandRegistry{}
	for _, c := range commands {
		cmdReg[c.Name] = c
	}
	return cmdReg
}

// getLogLevel converts string log levels to slog.Level representation.
// Can be one of "debug", "info", "warn" or "error".
func getLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
