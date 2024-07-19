package candy

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"github.com/lunixbochs/struc"
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
	net := getNetFromPath(c.Request.URL.Path)
	if net == nil {
		logger.Debug("net not found: %v", c.Request.URL.Path)
		return
	}
	ws := &candysocket{ctx: c, conn: conn, net: net}
	conn.SetPingHandler(func(buffer string) error { return ws.handlePingMessage(buffer) })

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
		switch uint8(buffer[0]) {
		case AUTH:
			err = ws.handleAuthMessage(buffer)
		case FORWARD:
			err = ws.handleForwardMessage(buffer)
		case DHCP:
			err = ws.handleDHCPMessage(buffer)
		case PEER:
			err = ws.handlePeerConnMessage(buffer)
		case VMAC:
			err = ws.handleVMacMessage(buffer)
		case DISCOVERY:
			err = ws.handleDiscoveryMessage(buffer)
		case GENERAL:
			err = ws.handleGeneralMessage(buffer)
		}
		if err != nil {
			logger.Debug("handle client message failed: %v", err)
			break
		}
	}

	if ws.dev != nil && ws.dev.model.Online {
		ws.dev.model.Online = false
		ws.dev.model.Save()

		net.ipWsMapMutex.Lock()
		defer net.ipWsMapMutex.Unlock()
		delete(net.ipWsMap, ws.dev.ip)
	}
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
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	return ws.conn.WriteMessage(websocket.CloseMessage, []byte(text))
}

func (ws *candysocket) writeMessage(buffer []byte) error {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	return ws.conn.WriteMessage(websocket.BinaryMessage, buffer)
}

func (ws *candysocket) writePong(buffer []byte) error {
	ws.connMutex.Lock()
	defer ws.connMutex.Unlock()
	return ws.conn.WriteMessage(websocket.PongMessage, buffer)
}

func (ws *candysocket) handlePingMessage(buffer string) error {
	ws.updateReadDeadline()

	err := func() error {
		ws.net.ipWsMapMutex.RLock()
		defer ws.net.ipWsMapMutex.RUnlock()

		if ws.dev.model == nil {
			return fmt.Errorf("ping failed: the client is not logged in: %v", buffer)
		}

		if ws.dev.model.Online {
			ws.dev.model.Save()
		}

		info := strings.Split(buffer, "::")
		if len(info) < 3 || info[0] != "candy" {
			return fmt.Errorf("ping failed: invalid format: %v", buffer)
		}

		ws.dev.model.OS = info[1]
		ws.dev.model.Version = info[2]

		if len(info) > 3 {
			ws.dev.model.Hostname = info[3]
		}

		return nil
	}()

	if err != nil {
		logger.Debug("client exception: %v", err)
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

	if ws.net.net != ws.net.mask&message.IP {
		return fmt.Errorf("auth failed: network does not match")
	}

	if ws.net.ipConflict(uint32ToStrIP(message.IP), ws.dev.model.VMac) {
		ws.writeCloseMessage("ip conflict")
		return fmt.Errorf("auth failed: ip conflict: %v", uint32ToStrIP(message.IP))
	}

	ws.net.ipWsMapMutex.Lock()
	defer ws.net.ipWsMapMutex.Unlock()

	if oldws, ok := ws.net.ipWsMap[message.IP]; ok {
		oldws.dev.model.Online = false
		oldws.dev.model.Save()
		oldws.writeCloseMessage("vmac conflict")
		oldws.conn.Close()
	}

	ws.dev.ip = message.IP
	ws.net.ipWsMap[message.IP] = ws

	db := storage.Get()
	db.Where(ws.dev.model).Find(ws.dev.model)
	ws.dev.model.IP = uint32ToStrIP(message.IP)
	ws.dev.model.Online = true
	ws.dev.model.Save()
	return nil
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

	if ws.dev.ip != message.Src {
		return fmt.Errorf("forward failed: source address does not match login information")
	}

	ws.dev.model.TX += uint64(len(buffer))

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
		dstWs.dev.model.RX += uint64(len(buffer))
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
		return false
	}()

	if broadcast {
		for _, dstWs := range ws.net.ipWsMap {
			if dstWs != ws && dstWs.dev.model.Online {
				dstWs.writeMessage(buffer)
				dstWs.dev.model.RX += uint64(len(buffer))
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

	if ws.net.model.DHCP == "" {
		return fmt.Errorf("dhcp failed: DHCP is not enabled")
	}

	cidr := func(input []byte) string {
		return string(input[:bytes.IndexByte(input[:], 0)])
	}(message.Cidr)

	if ws.dev.model == nil {
		return fmt.Errorf("dhcp failed: vmac not received")
	}
	db := storage.Get()
	ip, ipNet, err := net.ParseCIDR(cidr)
	needGenNewAddr := func() bool {
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
		db.Where(&model.Device{NetID: ws.net.model.ID, IP: ip.String()}).Find(&devices)
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

	var oldHost = ws.net.host
	for needGenNewAddr {

		result := db.Where(&model.Device{NetID: ws.net.model.ID, IP: ws.net.updateHost()})
		if result.RowsAffected == 0 {
			break
		}
		if oldHost == ws.net.host {
			return fmt.Errorf("dhcp failed: not enough addresses")
		}
	}

	if needGenNewAddr {
		ipNet := net.IPNet{
			IP:   make(net.IP, 4),
			Mask: make(net.IPMask, 4),
		}
		binary.BigEndian.PutUint32(ipNet.IP, ws.net.net|ws.net.host)
		binary.BigEndian.PutUint32(ipNet.Mask, ws.net.mask)
		message.Cidr = []byte(ipNet.String())
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

	return nil
}

func (ws *candysocket) handleVMacMessage(buffer []byte) error {
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

	ws.dev.model.TX += uint64(len(buffer))

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
		dstWs.dev.model.RX += uint64(len(buffer))
	}

	if uint32(0xFFFFFFFF) == message.Dst {
		for _, dstWs := range ws.net.ipWsMap {
			if dstWs != ws && dstWs.dev.model.Online {
				dstWs.writeMessage(buffer)
				dstWs.dev.model.RX += uint64(len(buffer))
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

	ws.dev.model.TX += uint64(len(buffer))

	ws.net.ipWsMapMutex.RLock()
	defer ws.net.ipWsMapMutex.RUnlock()

	if dstWs, ok := ws.net.ipWsMap[message.Dst]; ok {
		dstWs.writeMessage(buffer)
		dstWs.dev.model.RX += uint64(len(buffer))
	}

	if ws.net.model.Broadcast && uint32(0xFFFFFFFF) == message.Dst {
		for _, dstWs := range ws.net.ipWsMap {
			if dstWs != ws && dstWs.dev.model.Online {
				dstWs.writeMessage(buffer)
				dstWs.dev.model.RX += uint64(len(buffer))
			}
		}
	}

	return nil
}

func getNetFromPath(path string) *Net {
	result := strings.Split(strings.Trim(path, "/"), "/")
	if len(result) < 1 {
		return nil
	}
	username := result[0]
	netname := "@"
	if len(result) > 1 {
		if !IsAlphanumeric(result[1]) {
			return nil
		}
		netname = result[1]
	}
	netid := model.GetNetIdByUsernameAndNetname(username, netname)
	return getNetById(netid)
}
