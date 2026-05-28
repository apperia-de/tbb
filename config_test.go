package tbb_test

import (
	"os"
	"testing"

	"github.com/apperia-de/tbb"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestLoadConfig(t *testing.T) {
	t.Run("LoadConfig should panic because config file does not exists", func(t *testing.T) {
		assert.Panics(t, func() { tbb.LoadConfig("") })
	})

	t.Run("LoadConfig should panic because Telegram.BotToken is missing", func(t *testing.T) {
		assert.PanicsWithValue(t, "missing telegram bot token", func() { tbb.LoadConfig("test/data/test-missing-bot-token.config.yml") })
	})

	t.Run("LoadConfig should not panic when DB config is missing", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cfg := tbb.LoadConfig("test/data/test-missing-database.config.yml")
			assert.NotNil(t, cfg)
		})
	})

	t.Run("LoadConfig load as expected", func(t *testing.T) {
		cfg := tbb.LoadConfig("test/data/test.config.yml")
		assert.NotNil(t, cfg)
		assert.IsType(t, &tbb.Config{}, cfg)
	})
}

func TestCustomConfigExtension(t *testing.T) {
	type CustomConfig struct {
		tbb.Config `yaml:",inline"`
		Version    string `yaml:"version"`
		Username   string `yaml:"username"`
	}

	t.Run("Extend Config via composition and unmarshal directly", func(t *testing.T) {
		data, err := os.ReadFile("test/data/test.custom.config.yml")
		assert.NoError(t, err)

		var custom CustomConfig
		err = yaml.Unmarshal(data, &custom)
		assert.NoError(t, err)

		assert.Equal(t, "v0.1.0", custom.Version)
		assert.Equal(t, "me", custom.Username)
		assert.Equal(t, "123456:EXAMPLE", custom.Telegram.BotToken)
	})
}
