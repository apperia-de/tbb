package command

import (
	"fmt"
	"github.com/NicoNex/echotron/v3"
	"github.com/apperia-de/tbb"
)

type Timezone tbb.DefaultCommandHandler

func (c *Timezone) Handle(bot *tbb.Bot) tbb.StateFn {
	c.Bot = bot
	name := c.Bot.User().Firstname
	if name == "" {
		name = c.Bot.User().Username
	}
	_, _ = c.Bot.API().SendMessage(fmt.Sprintf("Hi %s, please send me a location in order to set the correct time zone for you.", name), c.Bot.ChatID(), nil)
	return c.awaitUserLocation
}

// awaitUserLocation waits for the user to send us the user's time zone
// and updates the timezone of the current user in the database.
func (c *Timezone) awaitUserLocation(u *echotron.Update) tbb.StateFn {
	if u.Message.Location == nil {
		_, _ = c.Bot.API().SendMessage("Please send a valid location point", u.ChatID(), nil)
		return c.awaitUserLocation
	}

	loc := *u.Message.Location
	_, _ = c.Bot.API().SendMessage(fmt.Sprintf("I receive your location update: Latitude = %f | Long = %f", loc.Latitude, loc.Longitude), u.ChatID(), nil)
	tzi := c.Bot.App().GetTimezoneInfo(loc.Latitude, loc.Longitude)

	user := c.Bot.User()
	user.UserInfo.Longitude = tzi.Longitude
	user.UserInfo.Latitude = tzi.Latitude
	user.UserInfo.Location = tzi.Location
	user.UserInfo.ZoneName = tzi.ZoneName
	user.UserInfo.IsDST = tzi.IsDST
	user.UserInfo.TZOffset = &tzi.Offset
	c.Bot.DB().Save(user)
	return nil
}
