package command

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/apperia-de/tbb"
)

type Timezone struct {
	tbb.DefaultCommandHandler
}

func (c *Timezone) Handle() tbb.StateFn {
	name := c.Bot().User().Firstname
	if name == "" {
		name = c.Bot().User().Username
	}
	_, _ = c.Bot().API().SendMessage(c.Bot().ChatID(), fmt.Sprintf("Hi %s, please send me a location in order to set the correct time zone for you.", name), nil)
	return c.awaitUserLocation
}

// awaitUserLocation waits for the user to send us the user's time zone
// and updates the timezone of the current user in the database.
func (c *Timezone) awaitUserLocation(u *gotgbot.Update) tbb.StateFn {
	if u.Message == nil || u.Message.Location == nil {
		chatID := c.Bot().ChatID()
		if u.Message != nil {
			chatID = u.Message.Chat.Id
		}
		_, _ = c.Bot().API().SendMessage(chatID, "Please send a valid location point", nil)
		return c.awaitUserLocation
	}

	loc := *u.Message.Location
	_, _ = c.Bot().API().SendMessage(u.Message.Chat.Id, fmt.Sprintf("I receive your location update: Latitude = %f | Long = %f", loc.Latitude, loc.Longitude), nil)
	tzi, err := c.Bot().TBot().GetTimezoneInfo(loc.Latitude, loc.Longitude)
	if err != nil {
		c.Bot().Log().Error("Error getting timezone info", "error", err)
		return nil
	}

	user := c.Bot().User()
	user.UserInfo.TimeZoneInfo = *tzi
	_ = c.Bot().Store().Save(user)
	return nil
}
