package tbb_test

import (
	"github.com/apperia-de/tbb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig should panic because config file does not exists", func(t *testing.T) {
		assert.Panics(t, func() { tbb.LoadConfig("") })
	})

	t.Run("LoadConfig should panic because Telegram.BotToken is missing", func(t *testing.T) {
		assert.PanicsWithValue(t, "missing telegram bot token", func() { tbb.LoadConfig("test/data/test-missing-bot-token.config.yml") })
	})

	t.Run("LoadConfig should panic because DB is missing", func(t *testing.T) {
		assert.PanicsWithValue(t, "missing database", func() { tbb.LoadConfig("test/data/test-missing-database.config.yml") })
	})

	t.Run("LoadConfig load as expected", func(t *testing.T) {
		cfg := tbb.LoadConfig("test/data/test.config.yml")
		assert.NotNil(t, cfg)
		assert.IsType(t, &tbb.Config{}, cfg)
	})
}

func TestLoadCustomConfig(t *testing.T) {
	type CustomConfig struct {
		Version   string `yaml:"version"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		Blacklist []int  `yaml:"blacklist"`
	}

	t.Run("LoadCustomConfig loads CustomConfig with missing CustomData", func(t *testing.T) {
		cfg := tbb.LoadCustomConfig[CustomConfig]("test/data/test.config.yml")
		assert.NotNil(t, cfg)
		assert.IsType(t, &tbb.Config{}, cfg)
		assert.NotNil(t, cfg.CustomData)
		assert.Equal(t, CustomConfig{}, cfg.CustomData)
	})

	t.Run("LoadCustomConfig loads CustomConfig with given CustomData", func(t *testing.T) {
		expected := CustomConfig{
			Version:   "v0.1.0",
			Username:  "me",
			Password:  "keins",
			Blacklist: []int{1, 2, 3},
		}

		cfg := tbb.LoadCustomConfig[CustomConfig]("test/data/test.custom.config.yml")
		assert.NotNil(t, cfg)
		assert.IsType(t, &tbb.Config{}, cfg)
		assert.IsType(t, CustomConfig{}, cfg.CustomData)
		assert.Equal(t, expected, cfg.CustomData)
	})

}
