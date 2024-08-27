package tbb

import (
	"embed"
	"encoding/json"
	"github.com/NicoNex/echotron/v3"
	timezone "github.com/evanoberholster/timezoneLookup/v2"
	"log/slog"
	"os"
	"strings"
)

//go:embed internal/data/timezone.data
var efs embed.FS

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

// GetLogLevel converts string log levels to slog.Level representation.
// Can be one of "debug", "info", "warn" or "error".
func GetLogLevel(level string) slog.Level {
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

func loadTimezoneCache() *timezone.Timezonecache {
	var (
		f, tempF *os.File
		tzc      timezone.Timezonecache
		data     []byte
		err      error
	)

	tempF, err = os.CreateTemp("", "timezone.data")
	if err != nil {
		panic(err)
	}
	defer os.Remove(tempF.Name())

	data, err = efs.ReadFile("internal/data/timezone.data")
	if err != nil {
		panic(err)
	}

	_, err = tempF.Write(data)
	if err != nil {
		panic(err)
	}

	f, err = os.Open(tempF.Name())
	if err != nil {
		panic(err)
	}

	if err = tzc.Load(f); err != nil {
		panic(err)
	}

	return &tzc
}
