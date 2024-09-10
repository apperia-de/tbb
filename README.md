# Telegram Bot Builder (tbb)

Tbb aims to provide a starting point for building Telegram bots in go.
The Telegram Bot Builder is based on the concurrent library [NicoNex/echotron](https://github.com/NicoNex/echotron).
To spin up a bot on your own see the examples section for details.

[![Go Report Card](https://goreportcard.com/badge/github.com/apperia-de/tbb)](https://goreportcard.com/report/github.com/apperia-de/tbb)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/apperia-de/tbb?style=flat)
![GitHub Licence](https://img.shields.io/github/license/apperia-de/tbb)

- Starting point for your own Telegram bot.
- Easily extendable.
- Implements Telegram bot user handling in either sqlite (default), mysql or postgres.
- Time zone handling by coordinates: Can use a location message from a user to set the current user time zone and offset from UTC.

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
	"github.com/apperia-de/tbb/command"
)

func main() {
	// Load your Telegram bot config (@see example.config.yml)
	cfg := tbb.LoadConfig('config.yml')
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
				Description: "Set your current time zone",
				Handler:     &command.Timezone{},
			},
			{
				Name:        "/help",
				Description: "Show the help message",
				Handler:     &command.Help{},
			},
		}),
	)
	
	app.Start() // Start a new bot polling for updates
}
```

### example.config.yml
```yaml
##############################################
# Telegram Bot Builder example configuration #
##############################################

debug: true
logLevel: info # One of debug | info | warn | error
telegram:
  botToken: "YOUR_TELEGRAM_BOT_TOKEN" # Enter your Telegram bot token which can be obtained from https://telegram.me/botfather
database:
  type: sqlite # One of sqlite | postgres | mysql
  filename: "app.db" # Only required for type sqlite
  #dsn: "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local" # Only required for type postgres or mysql
botSessionTimeout: 5 # Timeout in minutes before bot sessions will be deleted to save memory.
```

> For an example of how to implement your own UpdateHandler see `cmd/example/main.go` 