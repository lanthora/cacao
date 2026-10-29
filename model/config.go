package model

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type Config struct {
	gorm.Model
	Key   string `gorm:"index"`
	Value string
}

// All callers of setConfigValue are already inside an IMMEDIATE transaction.
// The lookup and write therefore remain serialized without a new unique index.
func configValue(db *gorm.DB, key, fallback string) (string, error) {
	var rows []Config
	if err := db.Where("key = ?", key).Limit(2).Find(&rows).Error; err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return fallback, nil
	}
	if len(rows) > 1 {
		return "", fmt.Errorf("%w: configuration key %q", ErrAmbiguousData, key)
	}
	return rows[0].Value, nil
}
func setConfigValue(tx *gorm.DB, key, value string) error {
	var rows []Config
	if err := tx.Where("key = ?", key).Limit(2).Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) > 1 {
		return fmt.Errorf("%w: configuration key %q", ErrAmbiguousData, key)
	}
	if len(rows) == 0 {
		return tx.Create(&Config{Key: key, Value: value}).Error
	}
	return changed(tx.Model(&Config{}).Where("id = ?", rows[0].ID).Update("value", value))
}
func GetConfig(key, fallback string, contexts ...context.Context) (string, error) {
	return configValue(readDB(contexts), key, fallback)
}
func SetConfig(key, value string, contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error { return setConfigValue(tx, key, value) })
}
func DelConfig(key string, contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error { return tx.Where("key = ?", key).Delete(&Config{}).Error })
}
