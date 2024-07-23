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
	if d.ID == 0 {
		db.Create(d)
	} else {
		db.Model(d).Select("*").Updates(d)
	}
}

func GetDevicesByNetID(netid uint) (devices []Device) {
	db := storage.Get()
	db.Where(&Device{NetID: netid}).Find(&devices)
	return
}

func GetDevicesByUserID(userid uint) (devices []Device) {
	db := storage.Get()
	db.Model(&Device{}).Joins("left join nets on devices.net_id = nets.id").Where("nets.user_id = ?", userid).Find(&devices)
	return
}
