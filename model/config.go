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

func SetString(key string, value string) {
	db := storage.Get()
	config := &Config{Key: key}
	db.Take(config)
	config.Value = value
	config.Save()
}

func GetString(key string) (value string, ok bool) {
	db := storage.Get()
	config := &Config{Key: key}

	if result := db.Where(config).Take(config); result.Error == nil {
		value = config.Value
		ok = true
	}

	return
}

func SetBool(key string, value bool) {
	SetString(key, func() string {
		if value {
			return "true"
		}
		return "false"
	}())
}

func GetBool(key string) (value bool, ok bool) {
	if v, ok := GetString(key); ok {
		return v == "true", true
	}
	return
}
