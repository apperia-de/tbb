package tbb

import (
	"encoding/json"
	"github.com/NicoNex/echotron/v3"
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

// GetUserFromUpdate returns the echotron.User from a given echotron.Update
func GetUserFromUpdate(u *echotron.Update) echotron.User {
	switch {
	case u.Message != nil:
		return *u.Message.From
	case u.EditedMessage != nil:
		return *u.EditedMessage.From
	case u.ChannelPost != nil:
		return *u.ChannelPost.From
	case u.EditedChannelPost != nil:
		return *u.EditedChannelPost.From
	case u.InlineQuery != nil:
		return *u.InlineQuery.From
	case u.ChosenInlineResult != nil:
		return *u.ChosenInlineResult.From
	case u.CallbackQuery != nil:
		return *u.CallbackQuery.From
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
	default:
		return echotron.User{ID: u.ChatID()}
	}
}

// GetChatTypeFromUpdate returns the ChatType from a given echotron.Update
func GetChatTypeFromUpdate(u *echotron.Update) ChatType {
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
func BuildInlineKeyboardButtonRow(buttons []InlineKeyboardButton) []echotron.InlineKeyboardButton {
	var res []echotron.InlineKeyboardButton
	for _, b := range buttons {
		res = append(res, echotron.InlineKeyboardButton{Text: b.Text, CallbackData: b.Data})
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
