package candy

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func init() {
	idNetMap = map[uint]*Net{}

	for _, netModel := range model.GetNets() {
		InsertNet(&netModel)
	}
}

type Net struct {
	model        *model.Net
	ipWsMap      map[uint32]*candysocket
	ipWsMapMutex sync.RWMutex

	net  uint32
	host uint32
	mask uint32
}

var idNetMap map[uint]*Net
var idNetMapMutex sync.RWMutex

func Flush() {
	idNetMapMutex.Lock()
	defer idNetMapMutex.Unlock()

	for _, n := range idNetMap {
		n.model.Update()
	}
}

func (net *Net) ipConflict(ip, vmac string) bool {
	db := storage.Get()
	device := &model.Device{NetID: net.model.ID, IP: ip}
	result := db.Where(device).Take(device)
	if result.Error == gorm.ErrRecordNotFound {
		return false
	}
	if result.Error == nil && device.VMac == vmac {
		return false
	}

	return true
}

func (net *Net) checkAuthMessage(message *AuthMessage) error {
	if absInt64(time.Now().Unix(), message.Timestamp) > 30 {
		return fmt.Errorf("auth check failed: timestamp: %v", message.Timestamp)
	}

	reported := message.Hash

	var data []byte
	data = append(data, net.model.Password...)
	data = binary.BigEndian.AppendUint32(data, message.IP)
	data = binary.BigEndian.AppendUint64(data, uint64(message.Timestamp))

	if sha256.Sum256([]byte(data)) != reported {
		return fmt.Errorf("auth check failed: hash does not match")
	}
	return nil
}

func (net *Net) checkDHCPMessage(message *DHCPMessage) error {
	if absInt64(time.Now().Unix(), message.Timestamp) > 30 {
		return fmt.Errorf("dhcp check failed: timestamp: %v", message.Timestamp)
	}

	reported := message.Hash

	var data []byte
	data = append(data, net.model.Password...)
	data = binary.BigEndian.AppendUint64(data, uint64(message.Timestamp))

	if sha256.Sum256([]byte(data)) != reported {
		return fmt.Errorf("dhcp check failed: hash does not match")
	}
	return nil
}

func (net *Net) checkVMacMessage(message *VMacMessage) error {
	if absInt64(time.Now().Unix(), message.Timestamp) > 30 {
		return fmt.Errorf("vmac check failed: timestamp: %v", message.Timestamp)
	}

	if _, err := strconv.ParseUint(message.VMac, 16, 64); err != nil {
		return fmt.Errorf("vmac check failed: invalid vmac")
	}

	reported := message.Hash

	var data []byte
	data = append(data, net.model.Password...)
	data = append(data, message.VMac...)
	data = binary.BigEndian.AppendUint64(data, uint64(message.Timestamp))

	if sha256.Sum256([]byte(data)) != reported {
		return fmt.Errorf("vmac check failed: hash does not match")
	}
	return nil
}

func (net *Net) updateHost() string {
	for ok := true; ok; ok = (net.host == 0 || net.host == ^net.mask) {
		net.host = (net.host + 1) & (^net.mask)
	}
	return uint32ToStrIP(net.net | net.host)
}

func (net *Net) Close() {
	net.ipWsMapMutex.Lock()
	defer net.ipWsMapMutex.Unlock()
	for ip, ws := range net.ipWsMap {
		ws.writeCloseMessage("net close")
		ws.conn.Close()
		delete(net.ipWsMap, ip)
	}
}

func IsInvalidDHCP(cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return true
	}
	mask := binary.BigEndian.Uint32(ipNet.Mask)
	return ^mask < 2
}

func InsertNet(netModel *model.Net) {
	idNetMapMutex.Lock()
	defer idNetMapMutex.Unlock()

	if IsInvalidDHCP(netModel.DHCP) {
		logger.Fatal("invalid net cidr: %v", netModel.DHCP)
		return
	}

	_, ipNet, _ := net.ParseCIDR(netModel.DHCP)
	netid := binary.BigEndian.Uint32(ipNet.IP)
	mask := binary.BigEndian.Uint32(ipNet.Mask)
	hostid := rand.Uint32() & ^mask

	net := &Net{
		model:   netModel,
		ipWsMap: make(map[uint32]*candysocket),
		net:     netid,
		host:    hostid,
		mask:    mask,
	}
	net.updateHost()
	idNetMap[netModel.ID] = net
}

func UpdateNet(netModel *model.Net) {
	DeleteNet(netModel.ID)
	InsertNet(netModel)
}

func DeleteNet(netid uint) {
	idNetMapMutex.Lock()
	defer idNetMapMutex.Unlock()

	if net, ok := idNetMap[netid]; ok {
		net.Close()
	}

	delete(idNetMap, netid)
}

func getNetById(netid uint) *Net {
	idNetMapMutex.RLock()
	defer idNetMapMutex.RUnlock()

	if net, ok := idNetMap[netid]; ok {
		return net
	}
	return nil
}

func absInt64(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

func uint32ToStrIP(ip uint32) string {
	var buffer []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, ip)
	return net.IP(buffer).String()
}
