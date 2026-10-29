package model

import (
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

const MaxGeoDBIntervalHours = 87600

// Serialize read-then-write transactions for SQLite and prevent duplicate keys.
var geoDBConfigWriteMu sync.Mutex

type GeoDBConfig struct {
	Type                string `json:"type"`
	DownloadURL         string `json:"downloadURL"`
	UpdateIntervalHours int    `json:"updateIntervalHours"`
	LastSuccessAt       string `json:"lastSuccessAt"`
}

func ValidGeoDBType(kind string) bool { return kind == "ipdb" || kind == "mmdb" }

func (c GeoDBConfig) Validate() error {
	if !ValidGeoDBType(c.Type) {
		return fmt.Errorf("unsupported database type")
	}
	if c.UpdateIntervalHours < 0 || c.UpdateIntervalHours > MaxGeoDBIntervalHours {
		return fmt.Errorf("update interval must be between 0 and %d hours", MaxGeoDBIntervalHours)
	}
	if c.DownloadURL != "" {
		u, err := url.Parse(c.DownloadURL)
		if err != nil || u.Hostname() == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Fragment != "" {
			return fmt.Errorf("download URL must be an HTTP(S) URL without credentials or a fragment")
		}
	}
	return nil
}

func GetGeoDBConfig(kind string) (GeoDBConfig, error) {
	c := GeoDBConfig{Type: kind, UpdateIntervalHours: 24}
	if !ValidGeoDBType(kind) {
		return c, fmt.Errorf("unsupported database type")
	}
	var rows []Config
	if err := storage.Get().Where("key IN ?", []string{kind + "DownloadURL", kind + "UpdateIntervalHours", kind + "LastSuccessAt"}).Find(&rows).Error; err != nil {
		return c, err
	}
	for _, row := range rows {
		switch row.Key {
		case kind + "DownloadURL":
			c.DownloadURL = row.Value
		case kind + "UpdateIntervalHours":
			// Invalid persisted settings must not silently enable a download.
			value, err := strconv.Atoi(row.Value)
			if err != nil {
				return c, err
			}
			c.UpdateIntervalHours = value
		case kind + "LastSuccessAt":
			c.LastSuccessAt = row.Value
		}
	}
	return c, c.Validate()
}

func setGeoDBValue(tx *gorm.DB, key, value string) error {
	var row Config
	result := tx.Where("key = ?", key).First(&row)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}
	row.Key, row.Value = key, value
	return tx.Save(&row).Error
}

func SaveGeoDBConfig(c GeoDBConfig) error {
	geoDBConfigWriteMu.Lock()
	defer geoDBConfigWriteMu.Unlock()
	if err := c.Validate(); err != nil {
		return err
	}
	return storage.Get().Transaction(func(tx *gorm.DB) error {
		if err := setGeoDBValue(tx, c.Type+"DownloadURL", c.DownloadURL); err != nil {
			return err
		}
		return setGeoDBValue(tx, c.Type+"UpdateIntervalHours", fmt.Sprint(c.UpdateIntervalHours))
	})
}

func SaveGeoDBSuccess(kind string, at time.Time) error {
	geoDBConfigWriteMu.Lock()
	defer geoDBConfigWriteMu.Unlock()
	if !ValidGeoDBType(kind) {
		return fmt.Errorf("unsupported database type")
	}
	return storage.Get().Transaction(func(tx *gorm.DB) error {
		return setGeoDBValue(tx, kind+"LastSuccessAt", at.UTC().Format(time.RFC3339Nano))
	})
}
