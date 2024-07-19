package model

import (
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func init() {
	db := storage.Get()
	err := db.AutoMigrate(User{})
	if err != nil {
		logger.Fatal("auto migrate users failed: %v", err)
	}
}

type User struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex"`
	Password string
	Token    string
	Role     string
	IP       string
}

func (u *User) Save() {
	db := storage.Get()
	db.Save(u)
}
