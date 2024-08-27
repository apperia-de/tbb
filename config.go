package tbb

import (
	"gopkg.in/yaml.v3"
	"os"
)

const (
	defaultSessionTimout = 15 // Default bot session timeout of 15 minutes of inactivity
)

type Config struct {
	Admin struct {
		BotToken string `yaml:"botToken"`
		ChatID   int64  `yaml:"chatID"`
	} `yaml:"admin"`
	Database struct {
		Type     string `yaml:"type"`     // one of sqlite, mysql, postgres
		DSN      string `yaml:"dsn"`      // in case of mysql or postgres
		Filename string `yaml:"filename"` // in case of sqlite
	} `yaml:"database"`
	Debug             bool   `yaml:"debug"`
	BotSessionTimeout int    `yaml:"botSessionTimeout"` // Timeout in minutes after which the bot instance will be deleted in order to save memory. Defaults to 15 minutes.
	LogLevel          string `yaml:"logLevel"`
	Telegram          struct {
		BotToken string `yaml:"botToken"`
	} `yaml:"telegram"`
}

func LoadConfig(filename string) Config {
	var cfg Config

	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		panic(err)
	}

	if cfg.Telegram.BotToken == "" {
		panic("missing telegram Bot token")
	}
	if cfg.Database.Filename == "" {
		panic("missing database")
	}
	if cfg.BotSessionTimeout == 0 {
		cfg.BotSessionTimeout = defaultSessionTimout
	}

	return cfg
}
