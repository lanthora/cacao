package model

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func requestContext(contexts []context.Context) context.Context {
	if len(contexts) > 0 && contexts[0] != nil {
		return contexts[0]
	}
	return context.Background()
}
func readDB(contexts []context.Context) *gorm.DB {
	return storage.Get().WithContext(requestContext(contexts))
}
func writeDB(contexts []context.Context, fn func(*gorm.DB) error) error {
	return storage.Transaction(requestContext(contexts), fn)
}
func changed(result *gorm.DB) error {
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

var ErrAmbiguousData = errors.New("existing duplicate records require review")

var ErrUsernameTaken = errors.New("username already taken")
var ErrNetworkExists = errors.New("network already exists")
var ErrRegistrationDisabled = errors.New("registration disabled")
var ErrRegisterTooOften = errors.New("register too often")
var ErrCredentialsChanged = errors.New("credentials changed; authenticate again")

// RegisterUser rechecks registration policy, first-admin selection and uniqueness
// in one IMMEDIATE transaction; the default network and registration switch commit
// atomically with the user. On retry, IDs assigned by rolled-back inserts are reset.
func RegisterUser(input User, defaultNet Net, adminCreated bool, contexts ...context.Context) (User, *Net, error) {
	var user User
	var network *Net
	err := writeDB(contexts, func(tx *gorm.DB) error {
		user = input
		network = nil
		if !adminCreated {
			open, err := configValue(tx, "openreg", "true")
			if err != nil {
				return err
			}
			if open != "true" {
				return ErrRegistrationDisabled
			}
			interval, err := configValue(tx, "reginterval", "1440")
			if err != nil {
				return err
			}
			minutes, err := strconv.ParseUint(interval, 10, 32)
			if err != nil || minutes > uint64((1<<63-1)/int64(time.Minute)) {
				return fmt.Errorf("invalid register interval %q", interval)
			}
			var recent int64
			if err := tx.Model(&User{}).Where("ip = ? AND role = ? AND created_at > ?", input.IP, "normal", time.Now().Add(-time.Duration(minutes)*time.Minute)).Count(&recent).Error; err != nil {
				return err
			}
			if recent > 0 {
				return ErrRegisterTooOften
			}
			var count int64
			if err := tx.Model(&User{}).Count(&count).Error; err != nil {
				return err
			}
			user.Role = "normal"
			if count == 0 {
				user.Role = "admin"
			}
		} else {
			user.Role = "normal"
		}
		var duplicates int64
		if err := tx.Model(&User{}).Where("name = ?", user.Name).Count(&duplicates).Error; err != nil {
			return err
		}
		if duplicates > 0 {
			return ErrUsernameTaken
		}
		if err := tx.Create(&user).Error; err != nil {
			if storage.IsUnique(err) {
				return ErrUsernameTaken
			}
			return err
		}
		if user.Role == "admin" {
			return setConfigValue(tx, "openreg", "false")
		}
		n := defaultNet
		n.UserID = user.ID
		if err := tx.Create(&n).Error; err != nil {
			return err
		}
		network = &n
		return nil
	})
	if err != nil {
		return User{}, nil, err
	}
	return user, network, nil
}

// DeleteUserTree is shared by administrative deletion and inactive-user cleanup.
// The optional cutoff is rechecked under the write lock, so recently active users
// are not deleted based on an earlier snapshot.
func DeleteUserTree(id uint, cutoff *time.Time, contexts ...context.Context) ([]uint, error) {
	var ids []uint
	err := writeDB(contexts, func(tx *gorm.DB) error {
		ids = nil
		var user User
		if err := tx.First(&user, id).Error; err != nil {
			return err
		}
		if cutoff != nil && (user.Role != "normal" || !user.UpdatedAt.Before(*cutoff)) {
			return nil
		}
		if err := tx.Model(&Net{}).Where("user_id = ?", id).Pluck("id", &ids).Error; err != nil {
			return err
		}
		if len(ids) > 0 {
			if err := tx.Where("net_id IN ?", ids).Delete(&Device{}).Error; err != nil {
				return err
			}
			if err := tx.Where("net_id IN ?", ids).Delete(&Route{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", ids).Delete(&Net{}).Error; err != nil {
				return err
			}
		}
		return changed(tx.Delete(&user))
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

type UserStats struct {
	NetNum uint64
	DevNum uint64
	RX     uint64
	TX     uint64
}

func GetUserStats(id uint, contexts ...context.Context) (UserStats, error) {
	var stats UserStats
	err := readDB(contexts).Raw(`SELECT
 (SELECT count(*) FROM nets WHERE user_id = ? AND deleted_at IS NULL) AS net_num,
 (SELECT count(*) FROM devices JOIN nets ON nets.id = devices.net_id WHERE nets.user_id = ? AND nets.deleted_at IS NULL AND devices.deleted_at IS NULL) AS dev_num,
 (SELECT coalesce(sum(rx),0) FROM devices JOIN nets ON nets.id = devices.net_id WHERE nets.user_id = ?) AS rx,
 (SELECT coalesce(sum(tx),0) FROM devices JOIN nets ON nets.id = devices.net_id WHERE nets.user_id = ?) AS tx`, id, id, id, id).Scan(&stats).Error
	return stats, err
}
