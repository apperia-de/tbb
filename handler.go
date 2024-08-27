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

type CommandFunc func(*Bot, Command, *echotron.Update) StateFn

type CommandHandler interface {
	Handle(*Bot) StateFn
}

type DefaultCommandHandler struct {
	Bot *Bot
}

func (h *DefaultCommandHandler) Handle(bot *Bot) StateFn {
	h.Bot = bot
	h.Bot.logger.Info("Command received!", "command", bot.cmd.Name, "params", fmt.Sprintf("[%s]", strings.Join(bot.cmd.Params, ",")))
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

func (d *DefaultUpdateHandler) Bot() *Bot {
	return d.bot
}

func (d *DefaultUpdateHandler) SetBot(bot *Bot) {
	d.bot = bot
}

func (d *DefaultUpdateHandler) HandleMessage(m echotron.Message) StateFn {
	d.bot.Log().Info("Method: HandleMessage", "Message", d.printAsJson(m))
	return nil
}

func (d *DefaultUpdateHandler) HandleEditedMessage(m echotron.Message) StateFn {
	d.bot.Log().Info("Method: HandleEditedMessage", "Message", d.printAsJson(m))
	return nil
}

func (d *DefaultUpdateHandler) HandleChannelPost(m echotron.Message) StateFn {
	d.bot.Log().Info("Method: HandleChannelPost", "Message", d.printAsJson(m))
	return nil
}

func (d *DefaultUpdateHandler) HandleEditedChannelPost(m echotron.Message) StateFn {
	d.bot.Log().Info("Method: HandleEditedChannelPost", "Message", d.printAsJson(m))
	return nil
}

func (d *DefaultUpdateHandler) HandleInlineQuery(i echotron.InlineQuery) StateFn {
	d.bot.Log().Info("Method: HandleInlineQuery", "InlineQuery", d.printAsJson(i))
	return nil
}

func (d *DefaultUpdateHandler) HandleChosenInlineResult(c echotron.ChosenInlineResult) StateFn {
	d.bot.Log().Info("Method: HandleChosenInlineResult", "ChosenInlineResult", d.printAsJson(c))
	return nil
}

func (d *DefaultUpdateHandler) HandleCallbackQuery(c echotron.CallbackQuery) StateFn {
	d.bot.Log().Info("Method: HandleCallbackQuery", "CallbackQuery", d.printAsJson(c))
	return nil
}

func (d *DefaultUpdateHandler) HandleShippingQuery(s echotron.ShippingQuery) StateFn {
	d.bot.Log().Info("Method: HandleShippingQuery", "ShippingQuery", d.printAsJson(s))
	return nil
}

func (d *DefaultUpdateHandler) HandlePreCheckoutQuery(p echotron.PreCheckoutQuery) StateFn {
	d.bot.Log().Info("Method: HandlePreCheckoutQuery", "PreCheckoutQuery", d.printAsJson(p))
	return nil
}

func (d *DefaultUpdateHandler) HandleChatMember(c echotron.ChatMemberUpdated) StateFn {
	d.bot.Log().Info("Method: HandleChatMember", "ChatMemberUpdated", d.printAsJson(c))
	return nil
}

func (d *DefaultUpdateHandler) HandleChatJoinRequest(c echotron.ChatJoinRequest) StateFn {
	d.bot.Log().Info("Method: HandleChatJoinRequest", "ChatJoinRequest", d.printAsJson(c))
	return nil
}

func (d *DefaultUpdateHandler) HandleMyChatMember(c echotron.ChatMemberUpdated) StateFn {
	d.bot.Log().Info("Method: HandleMyChatMember", "ChatMemberUpdated", d.printAsJson(c))

	status := c.NewChatMember.Status
	switch status {
	case memberStatusJoin:
		// User unblocked the Bot
		d.bot.Log().Info("Bot unblocked by user", "status", status, "user", d.Bot().user.Firstname)
		d.Bot().EnableUser()
	case memberStatusLeave:
		// User blocked the Bot
		d.bot.Log().Info("Bot blocked by user", "status", status, "user", d.Bot().user.Firstname)
		d.Bot().DisableUser()
	default:
		// Unknown
		d.bot.Log().Info("MyChatMember.Status", "status", status, "user", c.From)
	}

	return nil
}

func (d *DefaultUpdateHandler) printAsJson(v any) string {
	var (
		err     error
		jsonStr []byte
	)

	jsonStr, err = json.Marshal(v)
	if err != nil {
		d.bot.Log().Error(err.Error())
	}

	return string(jsonStr)
}
