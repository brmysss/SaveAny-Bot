package dao

import (
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/krau/SaveAny-Bot/config"
	"github.com/krau/SaveAny-Bot/logger"
	"github.com/krau/SaveAny-Bot/types"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

var db *gorm.DB

func Init() {
	if err := os.MkdirAll(filepath.Dir(config.Cfg.DB.Path), 755); err != nil {
		logger.L.Fatal("Failed to create data directory: ", err)
		os.Exit(1)
	}
	var err error
	db, err = gorm.Open(sqlite.Open(config.Cfg.DB.Path), &gorm.Config{
		Logger: glogger.New(logger.L, glogger.Config{
			Colorful:                  true,
			SlowThreshold:             time.Second * 5,
			LogLevel:                  glogger.Error,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		}),
		PrepareStmt: true,
	})
	if err != nil {
		logger.L.Fatal("Failed to open database: ", err)
		os.Exit(1)
	}
	logger.L.Debug("Database connected")
	db.AutoMigrate(&types.ReceivedFile{}, &types.User{})

	for _, admin := range config.Cfg.Telegram.Admins {
		CreateUser(int64(admin))
	}
}
