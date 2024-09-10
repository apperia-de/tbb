package tbb

import (
	"fmt"
	"github.com/apperia-de/tbb/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

const (
	DB_TYPE_SQLITE   = "sqlite"
	DB_TYPE_MYSQL    = "mysql"
	DB_TYPE_POSTGRES = "postgres"
)

// NewDB returns a new Database connection based on the given config files
func NewDB(cfg *Config, gormCfg *gorm.Config) *DB {
	var (
		gormDB *gorm.DB
		err    error
	)

	switch cfg.Database.Type {
	case DB_TYPE_SQLITE:
		if cfg.Database.Filename == "" {
			panic("database filename is required")
		}
		gormDB, err = gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?cache=shared", cfg.Database.Filename)), gormCfg)
	case DB_TYPE_MYSQL:
		if cfg.Database.DSN == "" {
			panic("database DSN is required")
		}
		gormDB, err = gorm.Open(mysql.Open(cfg.Database.DSN), gormCfg)
	case DB_TYPE_POSTGRES:
		if cfg.Database.DSN == "" {
			panic("database DSN is required")
		}
		gormDB, err = gorm.Open(postgres.Open(cfg.Database.DSN), gormCfg)
	default:
		panic(fmt.Sprintf("unsupported database type: %s", cfg.Database.Type))
	}

	if err != nil {
		panic(err)
	}

	return &DB{
		DB: gormDB,
	}
}

// FindUserByChatID return a user by Telegram chat id if exists or error otherwise.
func (db *DB) FindUserByChatID(chatID int64) (*model.User, error) {
	var (
		user model.User
		err  error
	)
	err = db.Preload("UserInfo").Preload("UserPhoto").First(&user, "chat_id = ?", chatID).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}
