package model

import (
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func init() {
	db := storage.Get()
	err := db.AutoMigrate(Net{})
	if err != nil {
		logger.Fatal("auto migrate nets failed: %v", err)
	}
}

type Net struct {
	gorm.Model
	UserID    uint   `gorm:"index:idx_net"`
	Name      string `gorm:"index:idx_net"`
	Password  string
	DHCP      string
	Broadcast bool
}

func (n *Net) Create() {
	db := storage.Get()
	db.Create(n)
}

func (n *Net) Update() {
	db := storage.Get()
	db.Model(n).Select("*").Updates(n)
}

func (n *Net) Delete() {
	db := storage.Get()
	db.Delete(n)
}

func GetNets() (nets []Net) {
	db := storage.Get()
	db.Find(&nets)
	return
}

func GetNetByNetID(netid uint) (net Net) {
	db := storage.Get()
	db.Where(&Net{Model: gorm.Model{ID: netid}}).Take(&net)
	return
}

func GetNetsByUserID(userid uint) (nets []Net) {
	db := storage.Get()
	db.Where(&Net{UserID: userid}).Find(&nets)
	return
}

func GetNetIdByUsernameAndNetname(username, netname string) uint {
	netid := uint(0)
	db := storage.Get()
	db.Model(&Net{}).Select("nets.id").Joins("left join users on users.id = nets.user_id").Where("users.name = ? and nets.name = ?", username, netname).Take(&netid)
	return netid
}

func DeleteNetByNetID(netid uint) {
	db := storage.Get()
	db.Delete(&Net{Model: gorm.Model{ID: netid}})
}
