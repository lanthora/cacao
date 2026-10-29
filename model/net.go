package model

import (
	"context"

	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

type Net struct {
	gorm.Model
	UserID    uint   `gorm:"index:idx_net"`
	Name      string `gorm:"index:idx_net"`
	Password  string
	DHCP      string
	Broadcast bool
	Lease     uint
}

func (n *Net) Create(contexts ...context.Context) error {
	original := *n
	var saved Net
	err := writeDB(contexts, func(tx *gorm.DB) error {
		saved = original
		var owner User
		if err := tx.First(&owner, saved.UserID).Error; err != nil {
			return err
		}
		if err := checkNetworkName(tx, saved.UserID, saved.Name, 0); err != nil {
			return err
		}
		return tx.Create(&saved).Error
	})
	if storage.IsUnique(err) {
		return ErrNetworkExists
	}
	if err == nil {
		*n = saved
	}
	return err
}
func (n *Net) Update(contexts ...context.Context) error {
	err := writeDB(contexts, func(tx *gorm.DB) error {
		if err := checkNetworkName(tx, n.UserID, n.Name, n.ID); err != nil {
			return err
		}
		return changed(tx.Model(&Net{}).Where("id = ? AND user_id = ?", n.ID, n.UserID).Updates(map[string]interface{}{"name": n.Name, "password": n.Password, "dhcp": n.DHCP, "broadcast": n.Broadcast, "lease": n.Lease}))
	})
	if storage.IsUnique(err) {
		return ErrNetworkExists
	}
	return err
}
func (n *Net) Delete(contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error {
		if err := changed(tx.Where("id = ? AND user_id = ?", n.ID, n.UserID).Delete(&Net{})); err != nil {
			return err
		}
		if err := tx.Where("net_id = ?", n.ID).Delete(&Device{}).Error; err != nil {
			return err
		}
		return tx.Where("net_id = ?", n.ID).Delete(&Route{}).Error
	})
}
func GetNets(contexts ...context.Context) ([]Net, error) {
	var rows []Net
	err := readDB(contexts).Find(&rows).Error
	return rows, err
}
func GetNetByNetID(id uint, contexts ...context.Context) (Net, error) {
	var row Net
	if id == 0 {
		return row, gorm.ErrRecordNotFound
	}
	err := readDB(contexts).First(&row, id).Error
	return row, err
}
func GetNetsByUserID(id uint, contexts ...context.Context) ([]Net, error) {
	var rows []Net
	err := readDB(contexts).Where("user_id = ?", id).Find(&rows).Error
	return rows, err
}
func GetNetIdByUsernameAndNetname(username, name string, contexts ...context.Context) (uint, error) {
	var rows []Net
	err := readDB(contexts).Select("nets.id").Joins("JOIN users ON users.id = nets.user_id AND users.deleted_at IS NULL").Where("users.name = ? AND nets.name = ?", username, name).Limit(2).Find(&rows).Error
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	if len(rows) > 1 {
		return 0, ErrAmbiguousData
	}
	return rows[0].ID, nil
}

func checkNetworkName(tx *gorm.DB, userID uint, name string, except uint) error {
	var count int64
	if err := tx.Model(&Net{}).Where("user_id = ? AND name = ? AND id <> ?", userID, name, except).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrNetworkExists
	}
	return nil
}
