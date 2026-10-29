package model

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"index"`
	Password string
	Token    string
	Role     string
	IP       string
}

func GetUsers(contexts ...context.Context) ([]User, error) {
	var rows []User
	err := readDB(contexts).Find(&rows).Error
	return rows, err
}
func FindUser(name string, contexts ...context.Context) (User, error) {
	var rows []User
	err := readDB(contexts).Where("name = ?", name).Limit(2).Find(&rows).Error
	if err != nil {
		return User{}, err
	}
	if len(rows) == 0 {
		return User{}, gorm.ErrRecordNotFound
	}
	if len(rows) > 1 {
		return User{}, ErrAmbiguousData
	}
	return rows[0], nil
}
func Authenticate(id uint, token string, contexts ...context.Context) (User, error) {
	var row User
	err := readDB(contexts).Where("id = ? AND token = ?", id, token).Take(&row).Error
	return row, err
}

// Conditional updates prevent stale login/password requests from overwriting a
// newer password or reviving a token invalidated by another request.
func (u *User) UpdateCredentials(password, token, ip string, contexts ...context.Context) error {
	old := *u
	err := writeDB(contexts, func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ? AND password = ? AND token = ?", old.ID, old.Password, old.Token).Updates(map[string]interface{}{"password": password, "token": token, "ip": ip})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrCredentialsChanged
		}
		return nil
	})
	if err == nil {
		u.Password = password
		u.Token = token
		u.IP = ip
	}
	return err
}
func RefreshUserLastActiveTimeByUserID(id uint, contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error {
		return changed(tx.Model(&User{}).Where("id = ?", id).UpdateColumn("updated_at", time.Now()))
	})
}
