package model

import (
	"testing"
	"time"

	"github.com/lanthora/cacao/storage"
)

func TestGeoDBConfigValidation(t *testing.T) {
	for _, kind := range []string{"ipdb", "mmdb"} {
		for _, url := range []string{"", "https://example.com/city?token=abc", "http://localhost/db"} {
			if err := (GeoDBConfig{Type: kind, DownloadURL: url, UpdateIntervalHours: 24}).Validate(); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, c := range []GeoDBConfig{
		{Type: "../ipdb"}, {Type: "mmdb", UpdateIntervalHours: -1}, {Type: "ipdb", UpdateIntervalHours: MaxGeoDBIntervalHours + 1},
		{Type: "ipdb", DownloadURL: "file:///tmp/db"}, {Type: "ipdb", DownloadURL: "https://"}, {Type: "ipdb", DownloadURL: "https://user:password@example.com/db"},
	} {
		if err := c.Validate(); err == nil {
			t.Fatalf("accepted invalid config: %+v", c)
		}
	}
}

func TestGeoDBConfigPersistence(t *testing.T) {
	// Preserve any pre-existing configuration when tests use the default storage.
	db := storage.Get()
	keys := []string{}
	for _, kind := range []string{"ipdb", "mmdb"} {
		for _, suffix := range []string{"DownloadURL", "UpdateIntervalHours", "LastSuccessAt"} {
			keys = append(keys, kind+suffix)
		}
	}
	var original []Config
	if err := db.Where("key IN ?", keys).Find(&original).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Unscoped().Where("key IN ?", keys).Delete(&Config{}).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Unscoped().Where("key IN ?", keys).Delete(&Config{})
		if len(original) > 0 {
			db.Create(&original)
		}
	})
	for _, kind := range []string{"ipdb", "mmdb"} {
		c, err := GetGeoDBConfig(kind)
		if err != nil || c.UpdateIntervalHours != 24 || c.DownloadURL != "" {
			t.Fatalf("defaults: %+v %v", c, err)
		}
		c.DownloadURL = "https://example.com/" + kind
		c.UpdateIntervalHours = 48
		if err := SaveGeoDBConfig(c); err != nil {
			t.Fatal(err)
		}
		at := time.Now().UTC()
		if err := SaveGeoDBSuccess(kind, at); err != nil {
			t.Fatal(err)
		}
		c.DownloadURL += "?new=1"
		c.LastSuccessAt = "forged"
		if err := SaveGeoDBConfig(c); err != nil {
			t.Fatal(err)
		}
		actual, err := GetGeoDBConfig(kind)
		if err != nil || actual.DownloadURL != c.DownloadURL || actual.UpdateIntervalHours != 48 || actual.LastSuccessAt != at.Format(time.RFC3339Nano) {
			t.Fatalf("persisted: %+v %v", actual, err)
		}
		var count int64
		db.Model(&Config{}).Where("key = ?", kind+"DownloadURL").Count(&count)
		if count != 1 {
			t.Fatalf("duplicate key rows: %d", count)
		}
	}
}
