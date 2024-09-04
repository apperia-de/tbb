package tbb

import (
	"github.com/NicoNex/echotron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBot_Update(t *testing.T) {
	t.Run("All users are allowed to use the bot if AllowedChatIDs is nil or empty", func(t *testing.T) {
		cfg := LoadConfig("test/data/test.config.yml")
		cfg.AllowedChatIDs = []int64{}
		app := NewApp(WithConfig(cfg))
		bot := app.newBot(99999999, app.logger.WithGroup("bot"), func() UpdateHandler { return &DefaultUpdateHandler{} })
		u := &echotron.Update{
			Message: &echotron.Message{
				Chat: echotron.Chat{
					Type:      "private",
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
					ID:        99999999,
				},
				Text: "/test_command",
			},
			ID: 123456,
		}
		bot.EnableUser()
		bot.user.UpdatedAt = time.Now()
		bot.Update(u)

		assert.Empty(t, cfg.AllowedChatIDs)
		assert.Nil(t, bot.Command())
		assert.True(t, bot.User().UserInfo.IsActive)
	})

	t.Run("Users which are not in AllowedChatIDs list cannot use the bot", func(t *testing.T) {
		cfg := LoadConfig("test/data/test.config.yml")
		cfg.AllowedChatIDs = []int64{12345678}
		commands := []Command{
			{
				Name:        "/test_command",
				Description: "",
				Handler:     &DefaultCommandHandler{},
			},
		}
		app := NewApp(WithConfig(cfg), WithCommands(commands))
		bot := app.newBot(99999999, app.logger.WithGroup("bot"), func() UpdateHandler { return &DefaultUpdateHandler{} })

		u := &echotron.Update{
			Message: &echotron.Message{
				Chat: echotron.Chat{
					Type:      "private",
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
					ID:        99999999,
				},
				Text: "/test_command",
			},
			ID: 123456,
		}
		bot.EnableUser()
		bot.user.UpdatedAt = time.Now()
		bot.Update(u)

		assert.Equal(t, []int64{12345678}, cfg.AllowedChatIDs)
		assert.Nil(t, bot.Command())
		assert.False(t, bot.User().UserInfo.IsActive)
	})

	t.Run("Users which are not in AllowedChatIDs list cannot use the bot", func(t *testing.T) {
		cfg := LoadConfig("test/data/test.config.yml")
		cfg.AllowedChatIDs = []int64{12345678, 99999999}
		commands := []Command{
			{
				Name:        "/test_command",
				Description: "",
				Handler:     &DefaultCommandHandler{},
			},
		}
		app := NewApp(WithConfig(cfg), WithCommands(commands))
		bot := app.newBot(99999999, app.logger.WithGroup("bot"), func() UpdateHandler { return &DefaultUpdateHandler{} })

		u := &echotron.Update{
			Message: &echotron.Message{
				Chat: echotron.Chat{
					Type:      "private",
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
					ID:        99999999,
				},
				Text: "/test_command",
			},
			ID: 123456,
		}
		bot.EnableUser()
		bot.user.UpdatedAt = time.Now()
		bot.Update(u)

		assert.Equal(t, []int64{12345678, 99999999}, cfg.AllowedChatIDs)
		assert.NotNil(t, bot.Command())
		assert.Equal(t, "/test_command", bot.Command().Name)
		assert.True(t, bot.User().UserInfo.IsActive)
	})
}
