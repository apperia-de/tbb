package main

import (
	"github.com/NicoNex/echotron/v3"
	"github.com/apperia-de/tbb"
	"github.com/apperia-de/tbb/pkg/command"
)

/*
 * To override the default tbb.DefaultUpdateHandler, you can create your own handler and implement only the methods you want to override.
 * The tbb.DefaultUpdateHandler handles all different kinds of Telegram updates and just logs them. The only update type which is handled different, is the HandleMyChatMember type.
 * The HandleMyChatMember update type is used to enable or disable the users of your bot. See tbb.DefaultUpdateHandler for details.
 */
type myBotHandler struct {
	tbb.DefaultUpdateHandler
}

func (h *myBotHandler) HandleMessage(m echotron.Message) tbb.StateFn {
	if m.Location != nil {
		tzi, err := h.Bot().App().GetTimezoneInfo(m.Location.Latitude, m.Location.Longitude)
		if err != nil {
			h.Bot().Log().Error(err.Error())
			return nil
		}
		_, _ = h.Bot().API().SendMessage("Got a location: ```json\n"+tbb.PrintAsJson(tzi, true)+"\n```", m.From.ID, &echotron.MessageOptions{ParseMode: echotron.MarkdownV2})
		return nil
	}
	_, _ = h.Bot().API().SendMessage("Echo: "+m.Text, m.From.ID, nil)
	return nil
}

func main() {
	// Load your Telegram bot config (@see example.config.yml)
	cfg := tbb.LoadConfig("config.yml")

	app := tbb.New(
		tbb.WithConfig(cfg),
		tbb.WithCommands([]tbb.Command{
			{
				Name:        "/start",
				Description: "",
				Handler:     &command.Enable{},
			}, {
				Name:        "/enable",
				Description: "Enable bot notifications",
				Handler:     &command.Enable{},
			},
			{
				Name:        "/disable",
				Description: "Disable bot notifications",
				Handler:     &command.Disable{},
			},
			{
				Name:        "/timezone",
				Description: "Set your current timezone",
				Handler:     &command.Timezone{},
			},
			{
				Name:        "/help",
				Description: "Show the help message",
				Handler:     &command.Help{},
			},
		}),
		tbb.WithHandlerFunc(func() tbb.UpdateHandler {
			return &myBotHandler{}
		}),
	)

	app.Start()
}
