package command

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/apperia-de/tbb"
)

type Enable struct {
	tbb.DefaultCommandHandler
}

func (c *Enable) Handle() tbb.StateFn {
	c.Bot().EnableUser()
	_ = c.Bot().Store().Save(c.Bot().User())

	if c.Bot().User().UserInfo.ZoneName == "" {
		var buttons [][]gotgbot.InlineKeyboardButton
		buttons = append(buttons, tbb.BuildInlineKeyboardButtonRow(
			[]tbb.InlineKeyboardButton{
				{Text: "Yes", Data: "update"},
				{Text: "No", Data: ""},
			}),
		)
		_, _ = c.Bot().API().SendMessage(c.Bot().ChatID(), "I don't have your current time zone for messaging. Do you want to send me your current location, so that I can figure out your current timezone settings?", &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: buttons}})
		return c.awaitUserAnswer
	}

	var buttons [][]gotgbot.InlineKeyboardButton
	buttons = append(buttons, tbb.BuildInlineKeyboardButtonRow(
		[]tbb.InlineKeyboardButton{
			{Text: "Yes", Data: "keep"},
			{Text: "No", Data: "update"},
		}),
	)
	userInfo := c.Bot().User().UserInfo
	_, _ = c.Bot().API().SendLocation(c.Bot().ChatID(), userInfo.Latitude, userInfo.Longitude, nil)
	_, _ = c.Bot().API().SendMessage(c.Bot().ChatID(), "Is this location still correct", &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: buttons}})
	return c.awaitUserAnswer
}

func (c *Enable) awaitUserAnswer(u *gotgbot.Update) (state tbb.StateFn) {
	if u.CallbackQuery == nil {
		return c.awaitUserAnswer
	}
	answer := u.CallbackQuery.Data
	c.Bot().Log().Info("Answer:" + answer)

	chatID := u.CallbackQuery.Message.GetChat().Id

	switch answer {
	case "update":
		_, _ = c.Bot().API().SendMessage(chatID, "Ok, so than please send me a valid location point", nil)
		state = c.awaitUserLocation
	case "keep":
		_, _ = c.Bot().API().SendMessage(chatID, fmt.Sprintf("Ok, then I'll keep your current time zone (%s)", c.Bot().User().UserInfo.ZoneName), nil)
	default:
		_, _ = c.Bot().API().SendMessage(chatID, "Ok, than I will use UTC timezone for your timezone. Your account is now enabled.", nil)
	}

	return state
}

func (c *Enable) awaitUserLocation(u *gotgbot.Update) tbb.StateFn {
	if u.Message != nil && u.Message.Location == nil {
		_, _ = c.Bot().API().SendMessage(u.Message.Chat.Id, "Please send a valid location point", nil)
		return c.awaitUserLocation
	}

	loc := *u.Message.Location
	_, _ = c.Bot().API().SendMessage(u.Message.Chat.Id, fmt.Sprintf("I receive your location update: Latitude = %f | Longitude = %f.\nYour notifications are now enabled.", loc.Latitude, loc.Longitude), nil)

	tzi, err := c.Bot().TBot().GetTimezoneInfo(loc.Latitude, loc.Longitude)
	if err != nil {
		c.Bot().Log().Error("Error getting timezone info", "error", err)
		return nil
	}

	user := c.Bot().User()
	user.UserInfo.Latitude = tzi.Latitude
	user.UserInfo.Longitude = tzi.Longitude
	user.UserInfo.Location = tzi.Location
	user.UserInfo.ZoneName = tzi.ZoneName
	user.UserInfo.IsDST = tzi.IsDST
	user.UserInfo.Offset = tzi.Offset
	_ = c.Bot().Store().Save(user)

	return nil
}
