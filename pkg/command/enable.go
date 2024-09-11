package command

import (
	"fmt"
	"github.com/NicoNex/echotron/v3"
	"github.com/apperia-de/tbb"
)

type Enable struct {
	tbb.DefaultCommandHandler
}

func (c *Enable) Handle() tbb.StateFn {
	c.Bot().EnableUser()
	c.Bot().DB().Save(c.Bot().User())

	if c.Bot().User().UserInfo.ZoneName == "" {
		var buttons [][]echotron.InlineKeyboardButton
		buttons = append(buttons, tbb.BuildInlineKeyboardButtonRow(
			[]tbb.InlineKeyboardButton{
				{Text: "Yes", Data: "update"},
				{Text: "No", Data: ""},
			}),
		)
		_, _ = c.Bot().API().SendMessage("I don't have your current time zone for messaging. Do you want to send me your current location, so that I can figure out your current timezone settings?", c.Bot().ChatID(), &echotron.MessageOptions{ReplyMarkup: &echotron.InlineKeyboardMarkup{InlineKeyboard: buttons}})
		return c.awaitUserAnswer
	}

	var buttons [][]echotron.InlineKeyboardButton
	buttons = append(buttons, tbb.BuildInlineKeyboardButtonRow(
		[]tbb.InlineKeyboardButton{
			{Text: "Yes", Data: "keep"},
			{Text: "No", Data: "update"},
		}),
	)
	userInfo := c.Bot().User().UserInfo
	_, _ = c.Bot().API().SendLocation(c.Bot().ChatID(), userInfo.Latitude, userInfo.Longitude, nil)
	_, _ = c.Bot().API().SendMessage("Is this location still correct", c.Bot().ChatID(), &echotron.MessageOptions{ReplyMarkup: &echotron.InlineKeyboardMarkup{InlineKeyboard: buttons}})
	return c.awaitUserAnswer
}

func (c *Enable) awaitUserAnswer(u *echotron.Update) (state tbb.StateFn) {
	if u.CallbackQuery == nil {
		return c.awaitUserAnswer
	}
	answer := u.CallbackQuery.Data
	c.Bot().Log().Info("Answer:" + answer)

	switch answer {
	case "update":
		_, _ = c.Bot().API().SendMessage("Ok, so than please send me a valid location point", u.ChatID(), nil)
		state = c.awaitUserLocation
	case "keep":
		_, _ = c.Bot().API().SendMessage(fmt.Sprintf("Ok, then I'll keep your current time zone (%s)", c.Bot().User().UserInfo.ZoneName), u.ChatID(), nil)
	default:
		_, _ = c.Bot().API().SendMessage("Ok, than I will use UTC timezone for your timezone. Your account is now enabled.", c.Bot().ChatID(), nil)
	}

	return state
}

func (c *Enable) awaitUserLocation(u *echotron.Update) tbb.StateFn {
	if u.Message != nil && u.Message.Location == nil {
		_, _ = c.Bot().API().SendMessage("Please send a valid location point", u.ChatID(), nil)
		return c.awaitUserLocation
	}

	loc := *u.Message.Location
	_, _ = c.Bot().API().SendMessage(fmt.Sprintf("I receive your location update: Latitude = %f | Longitude = %f.\nYour notifications are now enabled.", loc.Latitude, loc.Longitude), u.ChatID(), nil)

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
	c.Bot().DB().Save(user)

	return nil
}
