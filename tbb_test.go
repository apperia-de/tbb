package tbb_test

import (
	"github.com/apperia-de/tbb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("should create new tbot", func(t *testing.T) {
		cfg := tbb.LoadConfig("test/data/test.config.yml")
		tbot := tbb.New(tbb.WithConfig(cfg))
		if tbot == nil {
			t.Error("should return a new tbot")
		}
	})

	t.Run("should panic create new tbot without token", func(t *testing.T) {
		assert.Panics(t, func() { tbb.New() })
	})

	t.Run("should create new tbot with custom config", func(t *testing.T) {
		type CustomConfig struct {
			tbb.Config `yaml:",inline"`
			Version    string `yaml:"version"`
			Username   string `yaml:"username"`
		}

		customCfg := CustomConfig{
			Config: tbb.Config{
				Telegram: struct {
					BotToken string `yaml:"botToken"`
				}{BotToken: "123456:EXAMPLE"},
			},
			Version:  "v0.1.0",
			Username: "me",
		}

		tbot := tbb.New(tbb.WithConfig(&customCfg.Config))
		assert.NotNil(t, tbot)
	})
}

func ExampleNew() {
	type CustomConfig struct {
		tbb.Config `yaml:",inline"`
		Version    string   `yaml:"version"`
		Blacklist  []string `yaml:"blacklist"`
	}

	// Developer can load and parse their configuration directly:
	// data, _ := os.ReadFile("config.yml")
	// var cfg CustomConfig
	// yaml.Unmarshal(data, &cfg)

	cfg := CustomConfig{
		Config: tbb.Config{
			Telegram: struct {
				BotToken string `yaml:"botToken"`
			}{BotToken: "123456:YOUR_TELEGRAM_BOT_TOKEN"},
		},
	}

	tbot := tbb.New(tbb.WithConfig(&cfg.Config))
	tbot.Start()
}
