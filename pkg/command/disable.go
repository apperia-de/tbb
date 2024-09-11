package command

import "github.com/apperia-de/tbb"

type Disable struct {
	tbb.DefaultCommandHandler
}

func (c *Disable) Handle() tbb.StateFn {
	_, _ = c.Bot().API().SendMessage("You won't receive any updates anymore. Send /enable to enable updates again.", c.Bot().ChatID(), nil)
	c.Bot().DisableUser()
	c.Bot().DB().Save(c.Bot().User())
	return nil
}
