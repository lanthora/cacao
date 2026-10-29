package candy

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

type Net struct {
	stateMu          sync.RWMutex
	closed           bool
	addressRevisions map[string]uint64 // guarded by stateMu; invalidates pending DHCP offers
	model            *model.Net
	ipWsMap          map[uint32]*candysocket
	ipWsMapMutex     sync.RWMutex

	net  uint32
	host uint32
	mask uint32
}

var idNetMap = map[uint]*Net{}

// ConfigMu orders database commits and their runtime effects. Never acquire it inside a database transaction.
var ConfigMu sync.Mutex
var idNetMapMutex sync.RWMutex

func flush(contexts ...context.Context) {
	ctx := context.Background()
	if len(contexts) > 0 {
		ctx = contexts[0]
	}
	retryDisconnects(ctx)
	idNetMapMutex.RLock()
	nets := make([]*Net, 0, len(idNetMap))
	for _, n := range idNetMap {
		nets = append(nets, n)
	}
	idNetMapMutex.RUnlock()
	var devices []model.Device
	users := map[uint]bool{}
	for _, n := range nets {
		n.stateMu.RLock()
		n.ipWsMapMutex.RLock()
		for _, ws := range n.ipWsMap {
			if ws.dev.model.Online {
				devices = append(devices, model.Device{Model: ws.dev.model.Model, NetID: ws.dev.model.NetID, VMac: ws.dev.model.VMac, SessionID: ws.dev.model.SessionID, RX: atomic.LoadUint64(&ws.dev.model.RX), TX: atomic.LoadUint64(&ws.dev.model.TX)})
				users[n.model.UserID] = true
			}
		}
		n.ipWsMapMutex.RUnlock()
		n.stateMu.RUnlock()
	}
	// No application lock is held while waiting for SQLite.
	for _, d := range devices {
		if err := d.SaveTraffic(ctx); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("flush device %d: %v", d.ID, err)
		}
	}
	for id := range users {
		if err := model.RefreshUserLastActiveTimeByUserID(id, ctx); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("refresh user %d: %v", id, err)
		}
	}
}

func (n *Net) ipConflict(ctx context.Context, ip, vmac string) (bool, error) {
	var device model.Device
	err := storage.Get().WithContext(ctx).Where("net_id = ? AND ip = ?", n.model.ID, ip).Take(&device).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return device.VMac != vmac, nil
}

func (n *Net) checkAuthMessage(message *AuthMessage) error {
	reported := message.Hash

	var data []byte
	data = append(data, n.model.Password...)
	data = binary.BigEndian.AppendUint32(data, message.IP)
	data = binary.BigEndian.AppendUint64(data, uint64(message.Timestamp))

	if sha256.Sum256([]byte(data)) != reported {
		return fmt.Errorf("auth check failed: hash does not match")
	}
	return nil
}

func (n *Net) checkDHCPMessage(message *DHCPMessage) error {
	reported := message.Hash

	var data []byte
	data = append(data, n.model.Password...)
	data = binary.BigEndian.AppendUint64(data, uint64(message.Timestamp))

	if sha256.Sum256([]byte(data)) != reported {
		return fmt.Errorf("dhcp check failed: hash does not match")
	}
	return nil
}

func (n *Net) checkVMacMessage(message *VMacMessage) error {
	if _, err := strconv.ParseUint(message.VMac, 16, 64); err != nil {
		return fmt.Errorf("vmac check failed: invalid vmac")
	}

	reported := message.Hash

	var data []byte
	data = append(data, n.model.Password...)
	data = append(data, message.VMac...)
	data = binary.BigEndian.AppendUint64(data, uint64(message.Timestamp))

	if sha256.Sum256([]byte(data)) != reported {
		return fmt.Errorf("vmac check failed: hash does not match")
	}
	return nil
}

func (n *Net) updateHost() string {
	for ok := true; ok; ok = (n.host == 0 || n.host == ^n.mask) {
		n.host = (n.host + 1) & (^n.mask)
	}
	return uint32ToStrIp(n.net | n.host)
}

func (n *Net) close(retire bool) {
	n.stateMu.Lock()
	defer n.stateMu.Unlock()
	if retire {
		n.closed = true
	}
	n.ipWsMapMutex.Lock()
	defer n.ipWsMapMutex.Unlock()
	for ip, ws := range n.ipWsMap {
		ws.writeCloseMessage("net close")
		ws.conn.Close()
		delete(n.ipWsMap, ip)
	}
}

func IsInvalidDHCP(cidr string) bool {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil || ip.To4() == nil || len(ipNet.Mask) != net.IPv4len {
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
		net.close(true)
	}

	delete(idNetMap, netid)
}

func ReloadNet(netid uint) {
	idNetMapMutex.Lock()
	defer idNetMapMutex.Unlock()

	if net, ok := idNetMap[netid]; ok {
		net.close(false)
	}
}

func getNetById(netid uint) *Net {
	idNetMapMutex.RLock()
	defer idNetMapMutex.RUnlock()

	if net, ok := idNetMap[netid]; ok {
		return net
	}
	return nil
}

func getNetByPath(path string, contexts ...context.Context) *Net {
	username := "@"
	netname := "@"

	result := strings.Split(strings.Trim(path, "/"), "/")
	if IsValidUsername(result[0]) {
		username = result[0]
	}
	if len(result) > 1 {
		if !IsAlphaNumeric(result[1]) {
			return nil
		}
		netname = result[1]
	}
	netid, err := model.GetNetIdByUsernameAndNetname(username, netname, contexts...)
	if err != nil {
		logger.Warn("resolve network: %v", err)
		return nil
	}
	return getNetById(netid)
}

func absInt64(a, b int64) int64 {
	if a > b {
		return a - b
	}
	return b - a
}

func uint32ToStrIp(ip uint32) string {
	var buffer []byte = make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, ip)
	return net.IP(buffer).String()
}

func strIpToUint32(ip string) uint32 {
	segs := strings.Split(ip, ".")
	if len(segs) != 4 {
		logger.Fatal("convert ip to uint32 failed: %v", ip)
	}

	s := make([]int, 4)
	for i := 0; i < 4; i++ {
		n, err := strconv.Atoi(segs[i])
		if err != nil || n < 0 || n > 255 {
			logger.Fatal("convert ip to uint32 failed: %v", ip)
		}
		s[i] = n
	}

	rv := uint32(s[0]<<24 | s[1]<<16 | s[2]<<8 | s[3])
	return rv
}
