package model

import (
	"context"

	"gorm.io/gorm"
)

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

func (r *Route) Create(contexts ...context.Context) error {
	original := *r
	var saved Route
	err := writeDB(contexts, func(tx *gorm.DB) error {
		saved = original
		var parent Net
		if err := tx.First(&parent, original.NetID).Error; err != nil {
			return err
		}
		return tx.Create(&saved).Error
	})
	if err == nil {
		*r = saved
	}
	return err
}
func (r *Route) Delete(contexts ...context.Context) error {
	original := *r
	var saved Route
	err := writeDB(contexts, func(tx *gorm.DB) error { saved = original; return changed(tx.Delete(&saved)) })
	if err == nil {
		*r = saved
	}
	return err
}
func GetRouteByRouteID(id uint, contexts ...context.Context) (Route, error) {
	var row Route
	if id == 0 {
		return row, gorm.ErrRecordNotFound
	}
	err := readDB(contexts).First(&row, id).Error
	return row, err
}
func GetRoutesByUserID(id uint, contexts ...context.Context) ([]Route, error) {
	var rows []Route
	err := readDB(contexts).Joins("JOIN nets ON nets.id = routes.net_id AND nets.deleted_at IS NULL").Where("nets.user_id = ?", id).Order("routes.net_id,routes.priority").Find(&rows).Error
	return rows, err
}
