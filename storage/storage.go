package storage

import (
	"net"
	"os"
	"path"

	"github.com/glebarez/sqlite"
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/logger"
	"github.com/oschwald/geoip2-golang"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func init() {
	storageDir := argp.Get("storage", ".")
	err := os.MkdirAll(storageDir, os.ModeDir|os.ModePerm)
	if err != nil {
		logger.Fatal("make storage dir failed: %v", err)
	}
	db, err = gorm.Open(sqlite.Open(path.Join(storageDir, "sqlite.db")), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		logger.Fatal("open storage database failed: %v", err)
	}

	logger.Info("storage=[%v]", storageDir)
}

var db *gorm.DB

func Get() *gorm.DB {
	return db
}

func findFileByExtFromDir(dir string, ext string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if path.Ext(file.Name()) == ext {
			return file.Name(), nil
		}
	}
	return "", os.ErrNotExist
}

func GetCountryCity(ip string) (country, city string) {
	storageDir := argp.Get("storage", ".")
	filename, err := findFileByExtFromDir(storageDir, ".mmdb")
	if err != nil {
		logger.Debug("cannot found mmdb file: %v", err)
		return
	}
	db, err := geoip2.Open(path.Join(storageDir, filename))
	if err != nil {
		logger.Debug("open mmdb failed: %v", err)
		return
	}
	defer db.Close()
	record, err := db.City(net.ParseIP(ip))
	if err != nil {
		logger.Debug("get location mmdb failed: %v", err)
		return
	}
	country = record.Country.IsoCode
	city = record.City.Names["en"]
	return
}
