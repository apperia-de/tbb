# Telegram Bot TBot (tbb)

Tbb aims to provide a starting point for building Telegram bots in Go.
The Telegram Bot TBot is based on the modern, code-generated Telegram Bot API library [gotgbot/v2](https://github.com/PaulSonOfLars/gotgbot).
To spin up a bot on your own, see the examples section for details.


[![Go Report Card](https://goreportcard.com/badge/github.com/apperia-de/tbb)](https://goreportcard.com/report/github.com/apperia-de/tbb)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/apperia-de/tbb?style=flat)
![GitHub Licence](https://img.shields.io/github/license/apperia-de/tbb)

- Starting point for your own Telegram bot.
- Easily extendable.
- Database-agnostic: Implements a database-agnostic `UserStore` interface, defaulting to a zero-configuration thread-safe `InMemoryStore`.
- Public Commands: Allows specific commands to be marked as public, bypassing `AllowedChatIDs` checks.
- Time zone handling by coordinates: Can use the location message from the user to set the current user time zone and offset from UTC.

## How to use tbb

1. Create a new go project by `go mod init`.
2. Run `go get github.com/apperia-de/tbb`.
3. Create a new file `config.yml` with the contents from `example.config.yml`.
4. Adjust values to your needs, especially provide your **Telegram.BotToken**, which you may get from [@botfather](https://telegram.me/botfather) bot.
5. See example.

## Example

### Telegram bot

```go
package main

import (
	"github.com/apperia-de/tbb"
	"github.com/apperia-de/tbb/pkg/command"
)

func main() {
	// Load your Telegram bot config
	cfg := tbb.LoadConfig("config.yml")
	tbot := tbb.New(
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
				Description: "Set your current time zone",
				Handler:     &command.Timezone{},
			},
			{
				Name:        "/help",
				Description: "Show the help message",
				Handler:     &command.Help{},
			},
			{
				Name:        "/id",
				Description: "Get your Telegram user ID",
				Handler:     &command.ID{},
				Public:      true, // Bypasses AllowedChatIDs restrictions
			},
		}),
	)
	
	tbot.Start() // Start the new bot polling for updates
}
```

### example.config.yml
```yaml
##############################################
# Telegram Bot TBot example configuration    #
##############################################

debug: true
logLevel: info # One of debug | info | warn | error
telegram:
  botToken: "YOUR_TELEGRAM_BOT_TOKEN" # Enter your Telegram bot token which can be obtained from https://telegram.me/botfather
# Note: database type and connection parameters are omitted as we use default InMemoryStore
botSessionTimeout: 5 # Timeout in minutes before bot sessions will be deleted to save memory.
```

## How to use the newest `gotgbot` API

Since `tbb` exposes the raw `*gotgbot.Bot` client from `gotgbot/v2`, you can use any API method defined in the Telegram Bot API specification directly:

### 1. In a Command Handler
```go
func (c *MyCommand) Handle() tbb.StateFn {
    // Access the gotgbot Bot API client directly:
    bot := c.Bot().API()

    // Call SendMessage or any other standard API wrapper method:
    _, err := bot.SendMessage(c.Bot().ChatID(), "Hello world!", &gotgbot.SendMessageOpts{
        ParseMode: "HTML",
    })
    if err != nil {
        c.Bot().Log().Error("failed to send message", "error", err)
    }

    return nil
}
```

### 2. In a Custom Update Handler
```go
func (h *myBotHandler) HandleUpdate(u *gotgbot.Update) tbb.StateFn {
    if u.Message != nil {
        bot := h.Bot().API()
        // Echo back the message using raw API:
        _, _ = bot.SendMessage(u.Message.Chat.Id, "Echo: " + u.Message.Text, nil)
    }
    return nil
}
```

### 3. Using the `RouterUpdateHandler` Helper
If you prefer not to write type-switches, you can configure granular callbacks using the `RouterUpdateHandler` helper:
```go
app := tbb.New(
    tbb.WithConfig(cfg),
    tbb.WithHandlerFunc(func() tbb.UpdateHandler {
        return &tbb.RouterUpdateHandler{
            OnMessage: func(m *gotgbot.Message) tbb.StateFn {
                // Handle message update
                return nil
            },
            OnCallbackQuery: func(q *gotgbot.CallbackQuery) tbb.StateFn {
                // Handle callback query update
                return nil
            },
        }
    }),
)
```

> For an example of how to implement your own UpdateHandler see `cmd/example/main.go`