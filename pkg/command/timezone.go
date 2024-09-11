package command

import (
	"fmt"
	"github.com/NicoNex/echotron/v3"
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
	_, _ = c.Bot().API().SendMessage(fmt.Sprintf("Hi %s, please send me a location in order to set the correct time zone for you.", name), c.Bot().ChatID(), nil)
	return c.awaitUserLocation
}

// awaitUserLocation waits for the user to send us the user's time zone
// and updates the timezone of the current user in the database.
func (c *Timezone) awaitUserLocation(u *echotron.Update) tbb.StateFn {
	if u.Message.Location == nil {
		_, _ = c.Bot().API().SendMessage("Please send a valid location point", u.ChatID(), nil)
		return c.awaitUserLocation
	}

	loc := *u.Message.Location
	_, _ = c.Bot().API().SendMessage(fmt.Sprintf("I receive your location update: Latitude = %f | Long = %f", loc.Latitude, loc.Longitude), u.ChatID(), nil)
	tzi, err := c.Bot().TBot().GetTimezoneInfo(loc.Latitude, loc.Longitude)
	if err != nil {
		c.Bot().Log().Error("Error getting timezone info", "error", err)
		return nil
	}

	user := c.Bot().User()
	user.UserInfo.TimeZoneInfo = *tzi
	c.Bot().DB().Save(user)
	return nil
}
