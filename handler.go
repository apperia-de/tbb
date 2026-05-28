package tbb

import (
	"encoding/json"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"strings"
)

type CommandRegistry map[string]Command

type Command struct {
	Name        string
	Description string
	Params      []string
	Data        any
	Handler     CommandHandler
	Public      bool
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
	HandleUpdate(u *gotgbot.Update) StateFn
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

func (h *DefaultUpdateHandler) HandleUpdate(u *gotgbot.Update) StateFn {
	h.bot.Log().Info("Received update", "update", h.printAsJson(u))
	return nil
}

func (h *DefaultUpdateHandler) printAsJson(v any) string {
	jsonStr, _ := json.Marshal(v)
	return string(jsonStr)
}

// RouterUpdateHandler is a helper implementation of UpdateHandler that routes updates
// to individual configured callback functions.
type RouterUpdateHandler struct {
	bot *Bot

	OnMessage            func(*gotgbot.Message) StateFn
	OnEditedMessage      func(*gotgbot.Message) StateFn
	OnChannelPost        func(*gotgbot.Message) StateFn
	OnEditedChannelPost  func(*gotgbot.Message) StateFn
	OnInlineQuery        func(*gotgbot.InlineQuery) StateFn
	OnChosenInlineResult func(*gotgbot.ChosenInlineResult) StateFn
	OnCallbackQuery      func(*gotgbot.CallbackQuery) StateFn
	OnShippingQuery      func(*gotgbot.ShippingQuery) StateFn
	OnPreCheckoutQuery   func(*gotgbot.PreCheckoutQuery) StateFn
	OnPoll               func(*gotgbot.Poll) StateFn
	OnPollAnswer         func(*gotgbot.PollAnswer) StateFn
	OnMyChatMember       func(*gotgbot.ChatMemberUpdated) StateFn
	OnChatMember         func(*gotgbot.ChatMemberUpdated) StateFn
	OnChatJoinRequest    func(*gotgbot.ChatJoinRequest) StateFn
	OnChatBoost          func(*gotgbot.ChatBoostUpdated) StateFn
	OnRemovedChatBoost   func(*gotgbot.ChatBoostRemoved) StateFn
}

func (h *RouterUpdateHandler) Bot() *Bot {
	return h.bot
}

func (h *RouterUpdateHandler) SetBot(bot *Bot) {
	h.bot = bot
}

func (h *RouterUpdateHandler) HandleUpdate(u *gotgbot.Update) StateFn {
	switch {
	case u.Message != nil && h.OnMessage != nil:
		return h.OnMessage(u.Message)
	case u.EditedMessage != nil && h.OnEditedMessage != nil:
		return h.OnEditedMessage(u.EditedMessage)
	case u.ChannelPost != nil && h.OnChannelPost != nil:
		return h.OnChannelPost(u.ChannelPost)
	case u.EditedChannelPost != nil && h.OnEditedChannelPost != nil:
		return h.OnEditedChannelPost(u.EditedChannelPost)
	case u.InlineQuery != nil && h.OnInlineQuery != nil:
		return h.OnInlineQuery(u.InlineQuery)
	case u.ChosenInlineResult != nil && h.OnChosenInlineResult != nil:
		return h.OnChosenInlineResult(u.ChosenInlineResult)
	case u.CallbackQuery != nil && h.OnCallbackQuery != nil:
		return h.OnCallbackQuery(u.CallbackQuery)
	case u.ShippingQuery != nil && h.OnShippingQuery != nil:
		return h.OnShippingQuery(u.ShippingQuery)
	case u.PreCheckoutQuery != nil && h.OnPreCheckoutQuery != nil:
		return h.OnPreCheckoutQuery(u.PreCheckoutQuery)
	case u.Poll != nil && h.OnPoll != nil:
		return h.OnPoll(u.Poll)
	case u.PollAnswer != nil && h.OnPollAnswer != nil:
		return h.OnPollAnswer(u.PollAnswer)
	case u.MyChatMember != nil && h.OnMyChatMember != nil:
		return h.OnMyChatMember(u.MyChatMember)
	case u.ChatMember != nil && h.OnChatMember != nil:
		return h.OnChatMember(u.ChatMember)
	case u.ChatJoinRequest != nil && h.OnChatJoinRequest != nil:
		return h.OnChatJoinRequest(u.ChatJoinRequest)
	case u.ChatBoost != nil && h.OnChatBoost != nil:
		return h.OnChatBoost(u.ChatBoost)
	case u.RemovedChatBoost != nil && h.OnRemovedChatBoost != nil:
		return h.OnRemovedChatBoost(u.RemovedChatBoost)
	}
	return nil
}
