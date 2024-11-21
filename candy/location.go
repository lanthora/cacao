package candy

import (
	"crypto/tls"
	"net"
	"net/http"
	"path"

	"github.com/ipinfo/go/v2/ipinfo"
	"github.com/ipinfo/go/v2/ipinfo/cache"
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/util"
	"github.com/oschwald/geoip2-golang"
)

func GetLocation(ip net.IP) (country, region string) {
	if !ip.IsPrivate() {
		ok := false
		if country, region, ok = mmdbLocation(ip); !ok {
			country, region, _ = ipinfoLocation(ip)
		}
	}
	return
}

type dummyCacheEngine struct {
	cache map[string]interface{}
}

var dummyCache = ipinfo.NewCache(newDummyCacheEngine())
var httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

func newDummyCacheEngine() *dummyCacheEngine {
	return &dummyCacheEngine{
		cache: make(map[string]interface{}),
	}
}

func (c *dummyCacheEngine) Get(key string) (interface{}, error) {
	if v, ok := c.cache[key]; ok {
		return v, nil
	}
	return nil, cache.ErrNotFound
}

func (c *dummyCacheEngine) Set(key string, value interface{}) error {
	c.cache[key] = value
	return nil
}

func ipinfoLocation(ip net.IP) (country, region string, ok bool) {
	client := ipinfo.NewClient(httpClient, dummyCache, "")
	if info, err := client.GetIPInfo(ip); err == nil {
		country = info.Country
		region = info.Region
		ok = true
	}
	return
}

func mmdbLocation(ip net.IP) (country, region string, ok bool) {
	storageDir := argp.Get("storage", ".")
	filename, err := util.FindFileByExtFromDir(storageDir, ".mmdb")
	if err != nil {
		return
	}
	db, err := geoip2.Open(path.Join(storageDir, filename))
	if err != nil {
		logger.Debug("open mmdb failed: %v", err)
		return
	}
	defer db.Close()
	record, err := db.City(ip)
	if err != nil {
		logger.Debug("get location failed: %v", err)
		return
	}

	country = record.Country.IsoCode

	if len(record.Subdivisions) > 0 {
		region = record.Subdivisions[0].Names["en"]
	}
	ok = true
	return
}
