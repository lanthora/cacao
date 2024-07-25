package model

import (
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func init() {
	db := storage.Get()
	err := db.AutoMigrate(Route{})
	if err != nil {
		logger.Fatal("auto migrate routes failed: %v", err)
	}
}

type Route struct {
	gorm.Model
	NetID    uint `gorm:"index"`
	DevAddr  string
	DevMask  string
	DstAddr  string
	DstMask  string
	NextHop  string
	Priority int
}

func (r *Route) Create() {
	db := storage.Get()
	db.Create(r)
}

func (r *Route) Delete() {
	db := storage.Get()
	db.Delete(r)
}

func GetRouteByRouteID(routeid uint) (route Route) {
	if routeid != 0 {
		db := storage.Get()
		db.Where(&Route{Model: gorm.Model{ID: routeid}}).Take(&route)
	}
	return
}

func GetRoutesByUserID(userid uint) (routes []Route) {
	if userid != 0 {
		db := storage.Get()
		db.Model(&Route{}).Joins("left join nets on routes.net_id = nets.id").Where("nets.user_id = ?", userid).Order("routes.net_id,routes.priority").Find(&routes)
	}
	return
}
