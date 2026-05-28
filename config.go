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
		ChatIDs  []int64 `yaml:"chatIDs"`  // Telegram chat IDs of admins
	} `yaml:"admin"`
	AllowedChatIDs    []int64 `yaml:"allowedChatIDs"` // If set, only the specified chatIDs are allowed to use the bot. If not set or empty, all chat ids are allowed to use the bot.
	Debug             bool    `yaml:"debug"`
	BotSessionTimeout int     `yaml:"botSessionTimeout"` // Timeout in minutes, after which the bot instance will be deleted to save memory. Defaults to 15 minutes.
	LogLevel          string  `yaml:"logLevel"`
	Telegram          struct {
		BotToken string `yaml:"botToken"`
	} `yaml:"telegram"`
}

// LoadConfig returns the yaml config with the given name
func LoadConfig(filename string) *Config {
	return loadConfig(filename)
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
	if cfg.BotSessionTimeout == 0 {
		cfg.BotSessionTimeout = defaultSessionTimout
	}

	return &cfg
}
