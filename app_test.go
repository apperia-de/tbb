package tbb_test

import (
	"github.com/apperia-de/tbb"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewApp(t *testing.T) {
	t.Run("should create new app", func(t *testing.T) {
		cfg := tbb.LoadConfig("test/data/test.config.yml")
		app := tbb.NewApp(tbb.WithConfig(cfg))
		if app == nil {
			t.Error("should return a new app")
		}
		defer os.Remove("test/data/test.app.db")
	})

	t.Run("should panic create new app without config", func(t *testing.T) {
		assert.Panics(t, func() { tbb.NewApp() })
	})

	t.Run("should create new app with custom config", func(t *testing.T) {
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

		app := tbb.NewApp(tbb.WithConfig(customCfg))
		assert.NotNil(t, app)
		assert.Equal(t, expected, customCfg.CustomData)
	})
}

func ExampleNewApp() {
	type CustomConfig struct {
		Version   string   `yaml:"version"`
		Blacklist []string `yaml:"blacklist"`
	}

	cfg := tbb.LoadCustomConfig[CustomConfig]("config.yml")

	app := tbb.NewApp(tbb.WithConfig(cfg))
	app.Start()
}
