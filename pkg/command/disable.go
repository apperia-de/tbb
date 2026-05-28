package command

import "github.com/apperia-de/tbb"

type Disable struct {
	tbb.DefaultCommandHandler
}

func (c *Disable) Handle() tbb.StateFn {
	_, _ = c.Bot().API().SendMessage(c.Bot().ChatID(), "You won't receive any updates anymore. Send /enable to enable updates again.", nil)
	c.Bot().DisableUser()
	_ = c.Bot().Store().Save(c.Bot().User())
	return nil
}
