package command

import "github.com/apperia-de/tbb"

type Disable tbb.DefaultCommandHandler

func (c *Disable) Handle(bot *tbb.Bot) tbb.StateFn {
	_, _ = bot.API().SendMessage("You won't receive any updates anymore. Send /enable to enable updates again.", bot.ChatID(), nil)
	bot.DisableUser()
	return nil
}
