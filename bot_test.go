package tbb

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBot_Update(t *testing.T) {
	t.Run("All users are allowed to use the bot if AllowedChatIDs is nil or empty", func(t *testing.T) {
		cfg := LoadConfig("test/data/test.config.yml")
		cfg.AllowedChatIDs = []int64{}
		tbot := New(WithConfig(cfg))
		bot := tbot.newBot(99999999, tbot.logger.WithGroup("bot"), func() UpdateHandler { return &DefaultUpdateHandler{} })
		u := &gotgbot.Update{
			Message: &gotgbot.Message{
				Chat: gotgbot.Chat{
					Type:      "private",
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
					Id:        99999999,
				},
				From: &gotgbot.User{
					Id:        99999999,
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
				},
				Text: "/test_command",
			},
			UpdateId: 123456,
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
		tbot := New(WithConfig(cfg), WithCommands(commands))
		bot := tbot.newBot(99999999, tbot.logger.WithGroup("bot"), func() UpdateHandler { return &DefaultUpdateHandler{} })

		u := &gotgbot.Update{
			Message: &gotgbot.Message{
				Chat: gotgbot.Chat{
					Type:      "private",
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
					Id:        99999999,
				},
				From: &gotgbot.User{
					Id:        99999999,
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
				},
				Text: "/test_command",
			},
			UpdateId: 123456,
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
		tbot := New(WithConfig(cfg), WithCommands(commands))
		bot := tbot.newBot(99999999, tbot.logger.WithGroup("bot"), func() UpdateHandler { return &DefaultUpdateHandler{} })

		u := &gotgbot.Update{
			Message: &gotgbot.Message{
				Chat: gotgbot.Chat{
					Type:      "private",
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
					Id:        99999999,
				},
				From: &gotgbot.User{
					Id:        99999999,
					Username:  "test_user",
					FirstName: "test",
					LastName:  "user",
				},
				Text: "/test_command",
			},
			UpdateId: 123456,
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
