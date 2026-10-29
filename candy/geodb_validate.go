package candy

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ipipdotnet/ipdb-go"
	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

func validateGeoDB(kind, filename string) (err error) {
	// ipdb-go assumes trusted metadata in several slice operations.
	defer func() {
		if value := recover(); value != nil {
			err = fmt.Errorf("malformed database: %v", value)
		}
	}()
	switch kind {
	case "mmdb":
		db, err := maxminddb.Open(filename)
		if err != nil {
			return err
		}
		err = db.Verify()
		closeErr := db.Close()
		if err != nil {
			return err
		}
		if closeErr != nil {
			return closeErr
		}
		city, err := geoip2.Open(filename)
		if err != nil {
			return err
		}
		defer city.Close()
		_, err = city.City(net.ParseIP("1.1.1.1"))
		return err
	case "ipdb":
		return validateIPDB(filename)
	default:
		return errors.New("unsupported database type")
	}
}

func validateIPDB(filename string) error {
	body, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if len(body) < 4 {
		return errors.New("missing IPDB header")
	}
	headerSize := uint64(binary.BigEndian.Uint32(body[:4]))
	if headerSize > uint64(len(body)-4) {
		return errors.New("invalid IPDB metadata size")
	}
	var meta ipdb.MetaData
	if err := json.Unmarshal(body[4:4+int(headerSize)], &meta); err != nil {
		return err
	}
	data := body[4+int(headerSize):]
	if meta.TotalSize != len(data) || meta.NodeCount <= 0 || meta.NodeCount > len(data)/8 || meta.IPVersion&3 == 0 {
		return errors.New("invalid IPDB tree size or IP version")
	}
	language, ok := meta.Languages["CN"]
	if !ok || language < 0 || len(meta.Fields) == 0 {
		return errors.New("IPDB must support CN language")
	}
	fields := make(map[string]bool)
	for _, field := range meta.Fields {
		fields[field] = true
	}
	if !fields["country_code"] || (!fields["region_name"] && !fields["city_name"]) {
		return errors.New("IPDB must contain country_code and region_name or city_name")
	}
	// Validate every tree edge and referenced record before exposing this file.
	records := make(map[uint32]bool)
	for node := 0; node < meta.NodeCount; node++ {
		for side := 0; side < 2; side++ {
			offset := node*8 + side*4
			child := binary.BigEndian.Uint32(data[offset : offset+4])
			if uint64(child) < uint64(meta.NodeCount) {
				continue
			}
			if uint64(child) == uint64(meta.NodeCount) {
				return errors.New("IPDB contains an unsupported empty node")
			}
			if records[child] {
				continue
			}
			records[child] = true
			start := uint64(child) - uint64(meta.NodeCount) + uint64(meta.NodeCount)*8
			if start+2 > uint64(len(data)) {
				return errors.New("invalid IPDB record offset")
			}
			size := uint64(binary.BigEndian.Uint16(data[start : start+2]))
			if start+2+size > uint64(len(data)) {
				return errors.New("truncated IPDB record")
			}
			count := len(strings.Split(string(data[start+2:start+2+size]), "\t"))
			for _, langOffset := range meta.Languages {
				if langOffset < 0 || langOffset > count || len(meta.Fields) > count-langOffset {
					return errors.New("invalid IPDB record fields")
				}
			}
		}
	}
	db, err := ipdb.NewCity(filename)
	if err != nil {
		return err
	}
	probe := "1.1.1.1"
	if !db.IsIPv4() {
		probe = "2001:4860:4860::8888"
	}
	_, err = db.FindInfo(probe, "CN")
	if errors.Is(err, ipdb.ErrDataNotExists) {
		return nil
	}
	return err
}
