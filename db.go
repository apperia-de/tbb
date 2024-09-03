package tbb

import (
	"context"
	"fmt"
	"github.com/apperia-de/tbb/pkg/model"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log/slog"
)

type DB struct {
	*gorm.DB
	ctx    context.Context
	logger *slog.Logger
}

const (
	DB_TYPE_SQLITE   = "sqlite"
	DB_TYPE_MYSQL    = "mysql"
	DB_TYPE_POSTGRES = "postgres"
)

func newDB(cfg *Config, logger *slog.Logger) *DB {
	var (
		gormDB *gorm.DB
		err    error
	)

	gormCfg := gorm.Config{
		FullSaveAssociations: true,
	}

	switch cfg.Database.Type {
	case DB_TYPE_SQLITE:
		if cfg.Database.Filename == "" {
			panic("database filename is required")
		}
		gormDB, err = gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?cache=shared", cfg.Database.Filename)), &gormCfg)
	case DB_TYPE_MYSQL:
		if cfg.Database.DSN == "" {
			panic("database DSN is required")
		}
		gormDB, err = gorm.Open(mysql.Open(cfg.Database.DSN), &gormCfg)
	case DB_TYPE_POSTGRES:
		if cfg.Database.DSN == "" {
			panic("database DSN is required")
		}
		gormDB, err = gorm.Open(postgres.Open(cfg.Database.DSN), &gormCfg)
	default:
		panic(fmt.Sprintf("unsupported database type: %s", cfg.Database.Type))
	}

	if err != nil {
		panic(err)
	}

	db := &DB{
		DB:     gormDB,
		ctx:    context.Background(),
		logger: logger.WithGroup("database").With("debug", cfg.Debug, "verbose", getLogLevel(cfg.LogLevel) == slog.LevelDebug),
	}

	if err = db.AutoMigrate(&model.User{}, &model.UserInfo{}, &model.UserPhoto{}); err != nil {
		panic(err)
	}

	return db
}

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
