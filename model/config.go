package model

import (
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func init() {
	db := storage.Get()

	if err := db.AutoMigrate(Config{}); err != nil {
		logger.Fatal("auto migrate configs failed: %v", err)
	}
}

type Config struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex"`
	Value string
}

func (c *Config) Save() {
	db := storage.Get()
	db.Save(c)
}

func SetConfig(key string, value string) {
	db := storage.Get()
	config := &Config{Key: key}
	db.Where(config).Take(config)
	config.Value = value
	config.Save()
}

func GetConfig(key string, defaultValue string) string {
	db := storage.Get()
	config := &Config{Key: key}
	if result := db.Where(config).Take(config); result.Error == nil {
		return config.Value
	}
	return defaultValue
}

func DelConfig(key string) {
	db := storage.Get()
	config := &Config{Key: key}
	db.Where(config).Delete(config)
}
