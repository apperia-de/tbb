package tbb

import (
	"encoding/json"
	"fmt"
	"github.com/NicoNex/echotron/v3"
	"strings"
)

type CommandRegistry map[string]Command

type Command struct {
	Name        string
	Description string
	Params      []string
	Data        any
	Handler     CommandHandler
}

type CommandHandler interface {
	Bot() *Bot
	SetBot(*Bot)
	Handle() StateFn
}

type DefaultCommandHandler struct {
	bot *Bot
}

func (h *DefaultCommandHandler) Bot() *Bot {
	return h.bot
}

func (h *DefaultCommandHandler) SetBot(bot *Bot) {
	h.bot = bot
}

func (h *DefaultCommandHandler) Handle() StateFn {
	h.bot.logger.Info("Command received!", "command", h.bot.cmd.Name, "params", fmt.Sprintf("[%s]", strings.Join(h.bot.cmd.Params, ",")))
	return nil
}

type UpdateHandlerFn func() UpdateHandler

type UpdateHandler interface {
	Bot() *Bot
	SetBot(*Bot)
	HandleMessage(echotron.Message) StateFn
	HandleEditedMessage(echotron.Message) StateFn
	HandleChannelPost(echotron.Message) StateFn
	HandleEditedChannelPost(echotron.Message) StateFn
	HandleInlineQuery(echotron.InlineQuery) StateFn
	HandleChosenInlineResult(echotron.ChosenInlineResult) StateFn
	HandleCallbackQuery(echotron.CallbackQuery) StateFn
	HandleShippingQuery(echotron.ShippingQuery) StateFn
	HandlePreCheckoutQuery(echotron.PreCheckoutQuery) StateFn
	HandleChatMember(echotron.ChatMemberUpdated) StateFn
	HandleMyChatMember(echotron.ChatMemberUpdated) StateFn
	HandleChatJoinRequest(echotron.ChatJoinRequest) StateFn
}

// DefaultUpdateHandler implements the UpdateHandler interface
type DefaultUpdateHandler struct {
	bot *Bot
}

func (h *DefaultUpdateHandler) Bot() *Bot {
	return h.bot
}

func (h *DefaultUpdateHandler) SetBot(bot *Bot) {
	h.bot = bot
}

func (h *DefaultUpdateHandler) HandleMessage(m echotron.Message) StateFn {
	h.bot.Log().Info("Method: HandleMessage", "Message", h.printAsJson(m))
	return nil
}

func (h *DefaultUpdateHandler) HandleEditedMessage(m echotron.Message) StateFn {
	h.bot.Log().Info("Method: HandleEditedMessage", "Message", h.printAsJson(m))
	return nil
}

func (h *DefaultUpdateHandler) HandleChannelPost(m echotron.Message) StateFn {
	h.bot.Log().Info("Method: HandleChannelPost", "Message", h.printAsJson(m))
	return nil
}

func (h *DefaultUpdateHandler) HandleEditedChannelPost(m echotron.Message) StateFn {
	h.bot.Log().Info("Method: HandleEditedChannelPost", "Message", h.printAsJson(m))
	return nil
}

func (h *DefaultUpdateHandler) HandleInlineQuery(i echotron.InlineQuery) StateFn {
	h.bot.Log().Info("Method: HandleInlineQuery", "InlineQuery", h.printAsJson(i))
	return nil
}

func (h *DefaultUpdateHandler) HandleChosenInlineResult(c echotron.ChosenInlineResult) StateFn {
	h.bot.Log().Info("Method: HandleChosenInlineResult", "ChosenInlineResult", h.printAsJson(c))
	return nil
}

func (h *DefaultUpdateHandler) HandleCallbackQuery(c echotron.CallbackQuery) StateFn {
	h.bot.Log().Info("Method: HandleCallbackQuery", "CallbackQuery", h.printAsJson(c))
	return nil
}

func (h *DefaultUpdateHandler) HandleShippingQuery(s echotron.ShippingQuery) StateFn {
	h.bot.Log().Info("Method: HandleShippingQuery", "ShippingQuery", h.printAsJson(s))
	return nil
}

func (h *DefaultUpdateHandler) HandlePreCheckoutQuery(p echotron.PreCheckoutQuery) StateFn {
	h.bot.Log().Info("Method: HandlePreCheckoutQuery", "PreCheckoutQuery", h.printAsJson(p))
	return nil
}

func (h *DefaultUpdateHandler) HandleChatMember(c echotron.ChatMemberUpdated) StateFn {
	h.bot.Log().Info("Method: HandleChatMember", "ChatMemberUpdated", h.printAsJson(c))
	return nil
}

func (h *DefaultUpdateHandler) HandleChatJoinRequest(c echotron.ChatJoinRequest) StateFn {
	h.bot.Log().Info("Method: HandleChatJoinRequest", "ChatJoinRequest", h.printAsJson(c))
	return nil
}

func (h *DefaultUpdateHandler) HandleMyChatMember(c echotron.ChatMemberUpdated) StateFn {
	h.bot.Log().Info("Method: HandleMyChatMember", "ChatMemberUpdated", h.printAsJson(c))

	status := c.NewChatMember.Status
	switch status {
	case memberStatusJoin:
		// User unblocked the Bot
		h.bot.Log().Info("Bot unblocked by user", "status", status, "user", h.bot.user.Firstname)
		h.bot.EnableUser()
	case memberStatusLeave:
		// User blocked the Bot
		h.bot.Log().Info("Bot blocked by user", "status", status, "user", h.bot.user.Firstname)
		h.bot.DisableUser()
		h.bot.DB().Save(h.bot.user)
	default:
		// Unknown
		h.bot.Log().Info("MyChatMember.Status", "status", status, "user", c.From)
	}

	return nil
}

func (h *DefaultUpdateHandler) printAsJson(v any) string {
	var (
		err     error
		jsonStr []byte
	)

	jsonStr, err = json.Marshal(v)
	if err != nil {
		h.bot.Log().Error(err.Error())
	}

	return string(jsonStr)
}
