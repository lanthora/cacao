package candy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"github.com/lunixbochs/struc"
)

type deviceClient struct {
	t        *testing.T
	conn     *websocket.Conn
	password string
}

func deviceServer(t *testing.T, user model.User, network *model.Net) func(string) *deviceClient {
	t.Helper()
	router := gin.New()
	router.Use(WebsocketMiddleware())
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return func(vmac string) *deviceClient {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/"+user.Name, http.Header{"X-Forwarded-For": []string{"10.0.0.1"}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		client := &deviceClient{t, conn, network.Password}
		message := VMacMessage{Type: VMAC, VMac: vmac, Timestamp: time.Now().Unix()}
		payload := append([]byte(network.Password), []byte(vmac)...)
		payload = binary.BigEndian.AppendUint64(payload, uint64(message.Timestamp))
		message.Hash = sha256.Sum256(payload)
		client.send(&message)
		return client
	}
}

func (c *deviceClient) send(message interface{}) {
	c.t.Helper()
	var buffer bytes.Buffer
	if err := struc.Pack(&buffer, message); err != nil {
		c.t.Fatal(err)
	}
	if err := c.conn.WriteMessage(websocket.BinaryMessage, buffer.Bytes()); err != nil {
		c.t.Fatal(err)
	}
}
func (c *deviceClient) read() ([]byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, data, err := c.conn.ReadMessage()
	return data, err
}
func (c *deviceClient) dhcp(cidr string) string {
	c.t.Helper()
	message := DHCPMessage{Type: DHCP, Timestamp: time.Now().Unix(), Cidr: make([]byte, 32)}
	copy(message.Cidr, cidr)
	message.Hash = sha256.Sum256(binary.BigEndian.AppendUint64([]byte(c.password), uint64(message.Timestamp)))
	c.send(&message)
	data, err := c.read()
	if err != nil {
		c.t.Fatal(err)
	}
	var reply DHCPMessage
	if err := struc.Unpack(bytes.NewReader(data), &reply); err != nil {
		c.t.Fatal(err)
	}
	return string(bytes.TrimRight(reply.Cidr, "\x00"))
}
func (c *deviceClient) auth(ip string) {
	c.t.Helper()
	message := AuthMessage{Type: AUTH, IP: strIpToUint32(ip), Timestamp: time.Now().Unix()}
	payload := binary.BigEndian.AppendUint32([]byte(c.password), message.IP)
	payload = binary.BigEndian.AppendUint64(payload, uint64(message.Timestamp))
	message.Hash = sha256.Sum256(payload)
	c.send(&message)
}
func (c *deviceClient) echo(ip string) {
	c.t.Helper()
	c.send(&ForwardMessage{Type: FORWARD, Src: strIpToUint32(ip), Dst: strIpToUint32(ip)})
	if _, err := c.read(); err != nil {
		c.t.Fatal(err)
	}
}

func TestDeviceIPUpdateReconnectTrafficAndFailures(t *testing.T) {
	user, network := runtimeAccount(t, "ipeditlive")
	connect := deviceServer(t, user, network)
	first := connect("0000000000000011")
	ip := "192.168.202.10"
	if got := first.dhcp(ip + "/24"); got != ip+"/24" {
		t.Fatalf("initial DHCP: %s", got)
	}
	first.auth(ip)
	original := waitDevice(t, network.ID, func(d model.Device) bool { return d.IP == ip && d.Online })
	other := connect("0000000000000012")
	other.auth("192.168.202.11")
	waitDevice(t, network.ID, func(d model.Device) bool { return d.IP == "192.168.202.11" && d.Online })
	first.echo(ip)
	if _, err := UpdateDeviceIP(context.Background(), user.ID, original.ID, ip); err != nil {
		t.Fatal(err)
	}
	first.echo(ip) // unchanged address must leave the socket usable
	if _, err := UpdateDeviceIP(context.Background(), user.ID, original.ID, "192.168.202.11"); err != model.ErrDeviceIPConflict {
		t.Fatalf("online conflict: %v", err)
	}
	first.echo(ip)
	db := storage.Get()
	if err := db.Exec(fmt.Sprintf("CREATE TRIGGER reject_device_ip BEFORE UPDATE OF ip ON devices WHEN OLD.id = %d BEGIN SELECT RAISE(ABORT, 'injected IP failure'); END", original.ID)).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_device_ip") })
	if _, err := UpdateDeviceIP(context.Background(), user.ID, original.ID, "192.168.202.20"); err == nil {
		t.Fatal("database failure reported success")
	}
	first.echo(ip) // failed commit must not disconnect the client
	if err := db.Exec("DROP TRIGGER reject_device_ip").Error; err != nil {
		t.Fatal(err)
	}
	var workers sync.WaitGroup
	workers.Add(1)
	go func() {
		defer workers.Done()
		for i := 0; i < 10; i++ {
			flush()
		}
	}()
	if _, err := UpdateDeviceIP(context.Background(), user.ID, original.ID, "192.168.202.20"); err != nil {
		t.Fatal(err)
	}
	workers.Wait()
	if _, err := first.read(); !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
		t.Fatalf("expected targeted close: %v", err)
	}
	other.echo("192.168.202.11")
	saved, err := model.GetDeviceByDevID(original.ID)
	if err != nil || saved.Online || saved.IP != "192.168.202.20" || saved.RX != 4*21 || saved.TX != 4*21 {
		t.Fatalf("IP/traffic not saved atomically: %+v %v", saved, err)
	}
	next := connect("0000000000000011")
	if got := next.dhcp(ip + "/24"); got != "192.168.202.20/24" {
		t.Fatalf("database address not preferred: %s", got)
	}
	next.auth("192.168.202.20")
	replacement := waitDevice(t, network.ID, func(d model.Device) bool { return d.ID == original.ID && d.Online && d.SessionID != original.SessionID })
	original.Online = false
	persistDisconnect(original)
	fresh, err := model.GetDeviceByDevID(original.ID)
	if err != nil || !fresh.Online || fresh.IP != replacement.IP || fresh.SessionID != replacement.SessionID || fresh.RX != saved.RX {
		t.Fatalf("old disconnect overwrote new session: %+v %v", fresh, err)
	}
}

func TestDeviceIPUpdatePendingDHCPAndStaticClient(t *testing.T) {
	user, network := runtimeAccount(t, "ipeditpending")
	connect := deviceServer(t, user, network)
	device := model.Device{NetID: network.ID, VMac: "0000000000000021", IP: "192.168.202.30"}
	if err := storage.Get().Create(&device).Error; err != nil {
		t.Fatal(err)
	}
	pending := connect(device.VMac)
	if got := pending.dhcp("192.168.202.30/24"); got != "192.168.202.30/24" {
		t.Fatal(got)
	}
	if _, err := UpdateDeviceIP(context.Background(), user.ID, device.ID, "192.168.202.31"); err != nil {
		t.Fatal(err)
	}
	pending.auth("192.168.202.30")
	if _, err := pending.read(); err == nil {
		t.Fatal("stale DHCP offer was accepted")
	}
	fresh, err := model.GetDeviceByDevID(device.ID)
	if err != nil || fresh.IP != "192.168.202.31" || fresh.Online {
		t.Fatalf("pending login undid edit: %+v %v", fresh, err)
	}
	dynamic := connect(device.VMac)
	if got := dynamic.dhcp("192.168.202.30/24"); got != "192.168.202.31/24" {
		t.Fatal(got)
	}
	dynamic.auth("192.168.202.31")
	waitDevice(t, network.ID, func(d model.Device) bool { return d.ID == device.ID && d.Online })
	if _, err := UpdateDeviceIP(context.Background(), user.ID, device.ID, "192.168.202.32"); err != nil {
		t.Fatal(err)
	}
	static := connect(device.VMac)
	static.auth("192.168.202.30") // no DHCP request: client configuration remains authoritative
	waitDevice(t, network.ID, func(d model.Device) bool { return d.ID == device.ID && d.Online && d.IP == "192.168.202.30" })
	static.echo("192.168.202.30")
}
