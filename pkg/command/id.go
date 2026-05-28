package command

import (
	"fmt"
	"github.com/apperia-de/tbb"
)

type ID struct {
	tbb.DefaultCommandHandler
}

func (c *ID) Handle() tbb.StateFn {
	msg := fmt.Sprintf("Your Telegram Chat ID is: %d", c.Bot().ChatID())
	_, _ = c.Bot().API().SendMessage(msg, c.Bot().ChatID(), nil)
	return nil
}
