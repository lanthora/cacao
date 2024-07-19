package model

import (
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func init() {
	db := storage.Get()
	err := db.AutoMigrate(Device{})
	if err != nil {
		logger.Fatal("auto migrate devices failed: %v", err)
	}
}

type Device struct {
	gorm.Model
	NetID    uint
	VMac     string
	IP       string
	Online   bool
	RX       uint64
	TX       uint64
	OS       string
	Version  string
	Hostname string
}

func (d *Device) Save() {
	db := storage.Get()
	if result := db.Save(d); result.Error != nil {
		logger.Debug("save device failed: %v", result.Error)
	}
}
