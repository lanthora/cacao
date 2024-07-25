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
	db.Model(&Device{}).Where("online = true").Update("online", false)
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

func (d *Device) Delete() {
	db := storage.Get()
	db.Delete(d)
}

func GetDeviceByDevID(devid uint) (device Device) {
	if devid != 0 {
		db := storage.Get()
		db.Where(&Device{Model: gorm.Model{ID: devid}}).Find(&device)
	}
	return
}

func GetDevicesByNetID(netid uint) (devices []Device) {
	if netid != 0 {
		db := storage.Get()
		db.Where(&Device{NetID: netid}).Find(&devices)
	}
	return
}

func GetDevicesByUserID(userid uint) (devices []Device) {
	if userid != 0 {
		db := storage.Get()
		db.Model(&Device{}).Joins("left join nets on devices.net_id = nets.id").Where("nets.user_id = ?", userid).Find(&devices)
	}
	return
}

func GetRxSumByUserID(userid uint) (rx uint64) {
	if userid != 0 {
		db := storage.Get()
		db.Model(&Device{}).Select("sum(rx)").Joins("left join nets on devices.net_id = nets.id").Where("nets.user_id = ?", userid).Take(&rx)
	}
	return
}

func GetTxSumByUserID(userid uint) (tx uint64) {
	if userid != 0 {
		db := storage.Get()
		db.Model(&Device{}).Select("sum(tx)").Joins("left join nets on devices.net_id = nets.id").Where("nets.user_id = ?", userid).Take(&tx)
	}
	return
}

func DeleteDevicesByNetID(netid uint) (devices []Device) {
	if netid != 0 {
		db := storage.Get()
		db.Where(&Device{NetID: netid}).Delete(&Device{})
	}
	return
}
