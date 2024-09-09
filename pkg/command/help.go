package command

import (
	"fmt"
	"github.com/apperia-de/tbb"
)

const helpMessage = `Hi %s 👋. This is an example help message which describes what your bot can do.

Here is a list of available commands:

/enable  Enables the notifications
/disable Disables the notifications
/help    Shows this help message`

type Help struct {
	tbb.DefaultCommandHandler
}

// Handle function will be called on first command execution /start
func (c *Help) Handle() tbb.StateFn {
	name := c.Bot().User().Firstname
	if name == "" {
		name = c.Bot().User().Username
	}

	_, _ = c.Bot().API().SendMessage(fmt.Sprintf(helpMessage, name), c.Bot().ChatID(), nil)
	return nil
}
