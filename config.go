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
		BotToken string  `yaml:"botToken"` // Telegram bot token for an admin bot to use when sending messages
		ChatID   []int64 `yaml:"chatID"`   // Telegram chat IDs of admins
	} `yaml:"admin"`
	Database struct {
		Type     string `yaml:"type"`     // one of sqlite, mysql, postgres
		DSN      string `yaml:"dsn"`      // in the case of mysql or postgres
		Filename string `yaml:"filename"` // in the case of sqlite
	} `yaml:"database"`
	Debug             bool   `yaml:"debug"`
	BotSessionTimeout int    `yaml:"botSessionTimeout"` // Timeout in minutes after which the bot instance will be deleted in order to save memory. Defaults to 15 minutes.
	LogLevel          string `yaml:"logLevel"`
	Telegram          struct {
		BotToken string `yaml:"botToken"`
	} `yaml:"telegram"`
	CustomData any `yaml:"customData"`
}

// LoadConfig returns the yaml config with the given name
func LoadConfig(filename string) *Config {
	return loadConfig(filename)
}

// LoadCustomConfig returns the config but also takes your custom struct for the "customData" into account.
func LoadCustomConfig[T any](filename string) *Config {
	cfg := loadConfig(filename)
	out, err := yaml.Marshal(&cfg.CustomData)
	if err != nil {
		panic(err)
	}
	var custom T
	err = yaml.Unmarshal(out, &custom)
	if err != nil {
		panic(err)
	}
	cfg.CustomData = custom
	return cfg
}

func loadConfig(filename string) *Config {
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
		panic("missing telegram bot token")
	}

	if cfg.Database.Filename == "" {
		panic("missing database")
	}

	if cfg.BotSessionTimeout == 0 {
		cfg.BotSessionTimeout = defaultSessionTimout
	}

	return &cfg
}
