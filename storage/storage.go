package storage

import (
	"os"
	"path"

	"github.com/glebarez/sqlite"
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/logger"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {
	storageDir := argp.Get("storage", ".")
	err := os.MkdirAll(storageDir, os.ModeDir|os.ModePerm)
	if err != nil {
		logger.Fatal("make storage dir failed: %v", err)
	}
	db, err = gorm.Open(sqlite.Open(path.Join(storageDir, "sqlite.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.Fatal("open storage database failed: %v", err)
	}

	logger.Info("storage=[%v]", storageDir)
}

func Get() *gorm.DB {
	return db
}
