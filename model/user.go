package model

import (
	"time"

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
	Name     string `gorm:"index"`
	Password string
	Token    string
	Role     string
	IP       string
}

func (u *User) Save() {
	db := storage.Get()
	db.Save(u)
}

func (u *User) Delete() {
	db := storage.Get()
	db.Delete(u)
}

func GetUsers() (users []User) {
	db := storage.Get()
	db.Find(&users)
	return
}

func DeleteUserByUserID(userid uint) {
	db := storage.Get()
	db.Delete(&User{Model: gorm.Model{ID: userid}})
}

func GetLastActiveTimeByUserID(userid uint) (activeTime time.Time) {
	if userid != 0 {
		db := storage.Get()
		result := db.Unscoped().Model(&Device{}).Select("devices.updated_at").Joins("left join nets on devices.net_id = nets.id").Where("nets.user_id = ?", userid).Order("devices.updated_at desc").Take(&activeTime)
		if result.Error != nil {
			u := &User{Model: gorm.Model{ID: userid}}
			db.Model(u).Take(&u)
			activeTime = u.UpdatedAt
		}
	}
	return
}
