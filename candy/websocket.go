package candy

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"github.com/lunixbochs/struc"
	"gorm.io/gorm"
)

func WebsocketMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Upgrade") == "websocket" {
			handleWebsocket(c)
			c.Abort()
		} else {
			c.Next()
		}
	}
}

func handleWebsocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Debug("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()
	net := getNetByPath(c.Request.URL.Path, c.Request.Context())
	if net == nil {
		logger.Debug("net not found: %v", c.Request.URL.Path)
		return
	}
	ws := &candysocket{ctx: c, conn: conn, net: net}
	defer ws.disconnect()
	conn.SetPingHandler(func(buffer string) error {
		net.stateMu.Lock()
		defer net.stateMu.Unlock()
		if net.closed {
			return fmt.Errorf("network was removed or replaced")
		}
		return ws.handlePingMessage(buffer)
	})

	for {
		ws.updateReadDeadline()
		messageType, buffer, err := conn.ReadMessage()
		if err != nil {
			logger.Debug("read websocket failed: %v", err)
			break
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		err = ws.handleMessage(buffer)

		if err != nil {
			logger.Debug("handle client message failed: %v", err)
			break
		}
	}

}

func (ws *candysocket) disconnect() {
	net := ws.net
	net.stateMu.Lock()
	var disconnected *model.Device
	if ws.dev != nil && ws.dev.model.Online {
		ws.dev.model.Online = false
		snapshot := *ws.dev.model
		disconnected = &snapshot
		net.ipWsMapMutex.Lock()
		if net.ipWsMap[ws.dev.ip] == ws {
			delete(net.ipWsMap, ws.dev.ip)
		}
		net.ipWsMapMutex.Unlock()
	}
	net.stateMu.Unlock()
	if disconnected != nil {
		persistDisconnect(*disconnected)
	}

}

// Always release state locks on every return, including malformed-message panics.
func (ws *candysocket) handleMessage(buffer []byte) error {
	if len(buffer) == 0 {
		return nil
	}
	shared := buffer[0] == FORWARD || buffer[0] == GENERAL || buffer[0] == DISCOVERY
	if shared {
		ws.net.stateMu.RLock()
		defer ws.net.stateMu.RUnlock()
	} else {
		ws.net.stateMu.Lock()
		defer ws.net.stateMu.Unlock()
	}
	if ws.net.closed {
		return fmt.Errorf("network was removed or replaced")
	}
	switch buffer[0] {
	case AUTH:
		return ws.handleAuthMessage(buffer)
	case FORWARD:
		return ws.handleForwardMessage(buffer)
	case DHCP:
		return ws.handleDHCPMessage(buffer)
	case PEER:
		return ws.handlePeerConnMessage(buffer)
	case VMAC:
		return ws.handleVMacMessage(buffer)
	case DISCOVERY:
		return ws.handleDiscoveryMessage(buffer)
	case GENERAL:
		return ws.handleGeneralMessage(buffer)
	}
	return nil
}

type candysocket struct {
	ctx       *gin.Context
	conn      *websocket.Conn
	connMutex sync.Mutex
	dev       *Device
	net       *Net
}

func (ws *candysocket) updateReadDeadline() error {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	return ws.conn.SetReadDeadline((time.Now().Add(60 * time.Second)))
}

func (ws *candysocket) writeCloseMessage(text string) error {
	return ws.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, text), time.Now().Add(time.Second))
}

func (ws *candysocket) writeMessage(buffer []byte) error {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	if err := ws.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	return ws.conn.WriteMessage(websocket.BinaryMessage, buffer)
}

func (ws *candysocket) writePong(buffer []byte) error {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	if err := ws.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	return ws.conn.WriteMessage(websocket.PongMessage, buffer)
}

func (ws *candysocket) handlePingMessage(buffer string) error {
	ws.updateReadDeadline()

	if ws.dev == nil {
		logger.Debug("ping failed: the client is not logged in: %v", buffer)
		return nil
	}

	info := strings.Split(buffer, "::")
	if len(info) < 3 || info[0] != "candy" {
		logger.Debug("ping failed: invalid format: %v", buffer)
		return nil
	}

	ws.dev.model.OS = info[1]
	ws.dev.model.Version = info[2]

	if len(info) > 3 {
		ws.dev.model.Hostname = info[3]
	}

	if ws.dev.model.Online {
		if err := ws.dev.model.SaveOsVersionHostname(ws.ctx.Request.Context()); err != nil {
			return err
		}
	}

	ws.writePong([]byte(buffer))
	return nil
}

func (ws *candysocket) handleAuthMessage(buffer []byte) error {
	r := bytes.NewReader(buffer)
	message := &AuthMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	if err := ws.net.checkAuthMessage(message); err != nil {
		return err
	}

	if ws.dev == nil {
		return fmt.Errorf("auth failed: vmac not received")
	}

	if ws.net.net != ws.net.mask&message.IP || (^ws.net.mask)&(message.IP) == 0 || (^ws.net.mask)&(message.IP+1) == 0 {
		ws.writeCloseMessage("ip invalid")
		return fmt.Errorf("auth failed: network does not match")
	}

	conflict, err := ws.net.ipConflict(ws.ctx.Request.Context(), uint32ToStrIp(message.IP), ws.dev.model.VMac)
	if err != nil {
		return err
	}
	if conflict {
		ws.writeCloseMessage("ip conflict")
		return fmt.Errorf("auth failed: ip conflict: %v", uint32ToStrIp(message.IP))
	}

	ws.net.ipWsMapMutex.Lock()
	defer ws.net.ipWsMapMutex.Unlock()

	// Flush affected sessions before replacing them so unflushed traffic is kept.
	for _, oldws := range ws.net.ipWsMap {
		if oldws.dev.model.VMac == ws.dev.model.VMac || oldws.dev.ip == message.IP {
			if err := oldws.dev.model.SaveTraffic(ws.ctx.Request.Context()); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
	}
	candidate := *ws.dev.model
	candidate.IP = uint32ToStrIp(message.IP)
	candidate.Online = true
	candidate.SessionID = uuid.NewString()
	candidate.Country, candidate.Region = GetLocation(net.ParseIP(ws.ctx.ClientIP()))
	if err := candidate.Activate(ws.ctx.Request.Context()); err != nil {
		return err
	}
	if oldws, ok := ws.net.ipWsMap[message.IP]; ok && oldws != ws {
		oldws.dev.model.Online = false
		oldws.writeCloseMessage("vmac conflict")
		oldws.conn.Close()
	}
	// Replace other connections for the same device, including those on an old IP.
	for ip, oldws := range ws.net.ipWsMap {
		if oldws != ws && oldws.dev.model.ID == candidate.ID {
			oldws.dev.model.Online = false
			oldws.writeCloseMessage("vmac conflict")
			oldws.conn.Close()
			delete(ws.net.ipWsMap, ip)
		}
	}
	if ws.dev.model.Online {
		delete(ws.net.ipWsMap, ws.dev.ip)
	}
	*ws.dev.model = candidate
	ws.dev.ip = message.IP
	ws.net.ipWsMap[message.IP] = ws
	return ws.updateSystemRoute()
}

func (ws *candysocket) handleForwardMessage(buffer []byte) error {
	if ws.dev == nil {
		return fmt.Errorf("forward failed: conn is not logged in")
	}

	if !ws.dev.model.Online {
		return nil
	}

	r := bytes.NewReader(buffer)
	message := &ForwardMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	atomic.AddUint64(&ws.dev.model.TX, uint64(len(buffer)))

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
		atomic.AddUint64(&dstWs.dev.model.RX, uint64(len(buffer)))
	}

	broadcast := func() bool {
		if !ws.net.model.Broadcast {
			return false
		}
		if ws.net.net|^ws.net.mask == message.Dst {
			return true
		}
		if message.Dst&0xF0000000 == 0xE0000000 {
			return true
		}
		if message.Dst == 0xFFFFFFFF {
			return true
		}
		return false
	}()

	if broadcast {
		for _, dstWs := range ws.net.ipWsMap {
			if dstWs != ws && dstWs.dev.model.Online {
				dstWs.writeMessage(buffer)
				atomic.AddUint64(&dstWs.dev.model.RX, uint64(len(buffer)))
			}
		}
	}

	return nil
}

func (ws *candysocket) handleDHCPMessage(buffer []byte) error {
	r := bytes.NewReader(buffer)
	message := &DHCPMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	if err := ws.net.checkDHCPMessage(message); err != nil {
		return err
	}

	if ws.dev == nil || ws.dev.model == nil {
		return fmt.Errorf("dhcp failed: vmac not received")
	}

	db := storage.Get().WithContext(ws.ctx.Request.Context())

	var queryErr error
	// 检查能否使用数据库中地址, 地址可用时更新响应结果
	canUseLatestAddress := func() bool {
		device := model.Device{}
		if err := db.Where(&model.Device{NetID: ws.net.model.ID, VMac: ws.dev.model.VMac}).Take(&device).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				queryErr = err
			}
			return false
		}
		ip := net.ParseIP(device.IP)
		if ip == nil {
			return false
		}
		ip = ip.To4()
		if ip == nil {
			return false
		}

		if binary.BigEndian.Uint32(ip)&ws.net.mask != ws.net.net {
			return false
		}
		conflict, err := ws.net.ipConflict(ws.ctx.Request.Context(), device.IP, ws.dev.model.VMac)
		if err != nil {
			queryErr = err
			return false
		}
		if conflict {
			return false
		}

		ipNet := net.IPNet{IP: ip, Mask: make(net.IPMask, 4)}
		binary.BigEndian.PutUint32(ipNet.Mask, ws.net.mask)
		message.Cidr = []byte(ipNet.String())
		return true
	}()

	if queryErr != nil {
		return queryErr
	}
	// 根据用户传入的地址检查是否需要生成新地址
	needGenNewAddress := func() bool {
		if canUseLatestAddress {
			return false
		}
		cidr := func(input []byte) string {
			if end := bytes.IndexByte(input, 0); end >= 0 {
				return string(input[:end])
			}
			return string(input)
		}(message.Cidr)

		ip, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return true
		}
		if binary.BigEndian.Uint32(ipNet.IP) != ws.net.net {
			return true
		}
		if binary.BigEndian.Uint32(ipNet.Mask) != ws.net.mask {
			return true
		}
		devices := []model.Device{}
		if err := db.Where(&model.Device{NetID: ws.net.model.ID, IP: ip.String()}).Find(&devices).Error; err != nil {
			queryErr = err
			return true
		}
		if len(devices) > 1 {
			return true
		}
		if len(devices) == 0 {
			return false
		}
		if len(devices) == 1 && devices[0].VMac == ws.dev.model.VMac {
			return false
		}
		return true
	}()

	if queryErr != nil {
		return queryErr
	}
	// 用户传入的地址也不可用时, 生成一个新地址, 并更新响应结果
	var oldHost = ws.net.host
	for needGenNewAddress {
		device := &model.Device{NetID: ws.net.model.ID, IP: ws.net.updateHost()}
		err := db.Where(device).Take(&device).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ipNet := net.IPNet{IP: make(net.IP, 4), Mask: make(net.IPMask, 4)}
			binary.BigEndian.PutUint32(ipNet.IP, ws.net.net|ws.net.host)
			binary.BigEndian.PutUint32(ipNet.Mask, ws.net.mask)
			message.Cidr = []byte(ipNet.String())
			break
		}
		if oldHost == ws.net.host {
			ws.writeCloseMessage("not enough address")
			return fmt.Errorf("dhcp failed: not enough address")
		}
	}

	var output bytes.Buffer
	struc.Pack(&output, message)
	ws.writeMessage(output.Bytes())
	return nil
}

func (ws *candysocket) handlePeerConnMessage(buffer []byte) error {
	if ws.dev == nil {
		return fmt.Errorf("peer conn failed: conn is not logged in")
	}

	r := bytes.NewReader(buffer)
	message := &PeerConnMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	if ws.dev.ip != message.Src {
		return fmt.Errorf("peer conn failed: source address does not match login information")
	}

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
	}

	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, message.IP)
	ws.dev.model.Country, ws.dev.model.Region = GetLocation(ip)
	if err := ws.dev.model.SaveLocation(ws.ctx.Request.Context()); err != nil {
		return err
	}

	return nil
}

func (ws *candysocket) handleVMacMessage(buffer []byte) error {
	if ws.dev != nil && ws.dev.model.Online {
		return fmt.Errorf("cannot change device identity while online")
	}
	r := bytes.NewReader(buffer)
	message := &VMacMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	if err := ws.net.checkVMacMessage(message); err != nil {
		return err
	}
	ws.dev = &Device{model: &model.Device{NetID: ws.net.model.ID, VMac: message.VMac}}
	return nil
}

func (ws *candysocket) handleDiscoveryMessage(buffer []byte) error {
	if ws.dev == nil || !ws.dev.model.Online {
		return nil
	}

	r := bytes.NewReader(buffer)
	message := &DiscoveryMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	if ws.dev.ip != message.Src {
		return fmt.Errorf("discovery failed: source address does not match login information")
	}

	atomic.AddUint64(&ws.dev.model.TX, uint64(len(buffer)))

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
		atomic.AddUint64(&dstWs.dev.model.RX, uint64(len(buffer)))
	}

	if uint32(0xFFFFFFFF) == message.Dst {
		for _, dstWs := range ws.net.ipWsMap {
			if dstWs != ws && dstWs.dev.model.Online {
				dstWs.writeMessage(buffer)
				atomic.AddUint64(&dstWs.dev.model.RX, uint64(len(buffer)))
			}
		}
	}

	return nil
}

func (ws *candysocket) handleGeneralMessage(buffer []byte) error {
	if ws.dev == nil || !ws.dev.model.Online {
		return nil
	}

	r := bytes.NewReader(buffer)
	message := &GeneralMessage{}
	if err := struc.Unpack(r, message); err != nil {
		return err
	}

	if ws.dev.ip != message.Src {
		return fmt.Errorf("general failed: source address does not match login information")
	}

	atomic.AddUint64(&ws.dev.model.TX, uint64(len(buffer)))

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
		atomic.AddUint64(&dstWs.dev.model.RX, uint64(len(buffer)))
	}

	if ws.net.model.Broadcast && uint32(0xFFFFFFFF) == message.Dst {
		for _, dstWs := range ws.net.ipWsMap {
			if dstWs != ws && dstWs.dev.model.Online {
				dstWs.writeMessage(buffer)
				atomic.AddUint64(&dstWs.dev.model.RX, uint64(len(buffer)))
			}
		}
	}

	return nil
}

func (ws *candysocket) updateSystemRoute() error {
	header := &RouteMessage{Type: ROUTE, Size: 0, Reserved: 0}
	bodyBuffer := bytes.Buffer{}

	db := storage.Get().WithContext(ws.ctx.Request.Context())
	routes := []model.Route{}
	if err := db.Where(&model.Route{NetID: ws.net.model.ID}).Order("priority").Find(&routes).Error; err != nil {
		return err
	}
	for _, route := range routes {
		deviceAddr := strIpToUint32(route.DevAddr)
		deviceMask := strIpToUint32(route.DevMask)
		if deviceAddr != deviceMask&ws.dev.ip {
			continue
		}
		header.Size += 1
		destAddr := strIpToUint32(route.DstAddr)
		destMask := strIpToUint32(route.DstMask)
		nextHop := strIpToUint32(route.NextHop)
		body := &RouteMessageEntry{Dest: destAddr, Mask: destMask, NextHop: nextHop}
		struc.Pack(&bodyBuffer, body)
	}
	if header.Size > 0 {
		headerBuffer := bytes.Buffer{}
		struc.Pack(&headerBuffer, header)
		ws.writeMessage(append(headerBuffer.Bytes(), bodyBuffer.Bytes()...))
	}
	return nil
}
