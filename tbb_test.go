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

	t.Run("should panic create new tbot without config", func(t *testing.T) {
		assert.Panics(t, func() { tbb.New() })
	})

	t.Run("should create new tbot with custom config", func(t *testing.T) {
		type CustomConfig struct {
			Version   string `yaml:"version"`
			Username  string `yaml:"username"`
			Password  string `yaml:"password"`
			Blacklist []int  `yaml:"blacklist"`
		}

		expected := CustomConfig{
			Version:   "v0.1.0",
			Username:  "me",
			Password:  "keins",
			Blacklist: []int{1, 2, 3},
		}

		customCfg := tbb.LoadCustomConfig[CustomConfig]("test/data/test.custom.config.yml")

		tbot := tbb.New(tbb.WithConfig(customCfg))
		assert.NotNil(t, tbot)
		assert.Equal(t, expected, customCfg.CustomData)
	})
}

func ExampleNew() {
	type CustomConfig struct {
		Version   string   `yaml:"version"`
		Blacklist []string `yaml:"blacklist"`
	}

	cfg := tbb.LoadCustomConfig[CustomConfig]("config.yml")

	tbot := tbb.New(tbb.WithConfig(cfg))
	tbot.Start()
}
