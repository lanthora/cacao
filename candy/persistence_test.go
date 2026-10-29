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
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"github.com/lunixbochs/struc"
)

func runtimeAccount(t *testing.T, name string) (model.User, *model.Net) {
	t.Helper()
	user, network, err := model.RegisterUser(model.User{Name: name, Token: name}, model.Net{Name: "@", Password: "secret", DHCP: "192.168.202.0/24", Broadcast: true}, true)
	if err != nil {
		t.Fatal(err)
	}
	InsertNet(network)
	t.Cleanup(func() {
		DeleteNet(network.ID)
		if _, err := model.DeleteUserTree(user.ID, nil); err != nil {
			t.Error(err)
		}
	})
	return user, network
}
func waitDevice(t *testing.T, network uint, condition func(model.Device) bool) model.Device {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		devices, err := model.GetDevicesByNetID(network)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range devices {
			if condition(d) {
				return d
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("device state did not converge")
	return model.Device{}
}
func TestSocketTrafficFlushReconnectAndDisconnect(t *testing.T) {
	user, network := runtimeAccount(t, "sockettest")
	router := gin.New()
	router.Use(WebsocketMiddleware())
	server := httptest.NewServer(router)
	defer server.Close()
	connect := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/"+user.Name, http.Header{"X-Forwarded-For": []string{"10.0.0.1"}})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { conn.Close() })
		send := func(message interface{}) {
			var buffer bytes.Buffer
			if err := struc.Pack(&buffer, message); err != nil {
				t.Fatal(err)
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buffer.Bytes()); err != nil {
				t.Fatal(err)
			}
		}
		vmac := VMacMessage{Type: VMAC, VMac: "0000000000000001", Timestamp: time.Now().Unix()}
		var timestamp [8]byte
		binary.BigEndian.PutUint64(timestamp[:], uint64(vmac.Timestamp))
		vmac.Hash = sha256.Sum256(append(append([]byte(network.Password), []byte(vmac.VMac)...), timestamp[:]...))
		send(&vmac)
		auth := AuthMessage{Type: AUTH, IP: 0xc0a8ca01, Timestamp: time.Now().Unix()}
		var payload bytes.Buffer
		payload.WriteString(network.Password)
		_ = binary.Write(&payload, binary.BigEndian, auth.IP)
		_ = binary.Write(&payload, binary.BigEndian, auth.Timestamp)
		auth.Hash = sha256.Sum256(payload.Bytes())
		send(&auth)
		return conn
	}
	first := connect()
	original := waitDevice(t, network.ID, func(d model.Device) bool { return d.Online })
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ {
			flush()
		}
	}()
	var packed bytes.Buffer
	message := ForwardMessage{Type: FORWARD, Src: 0xc0a8ca01, Dst: 0xc0a8ca01}
	if err := struc.Pack(&packed, &message); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 30; i++ {
		if err := first.WriteMessage(websocket.BinaryMessage, packed.Bytes()); err != nil {
			t.Fatal(err)
		}
		first.SetReadDeadline(time.Now().Add(3 * time.Second))
		if _, _, err := first.ReadMessage(); err != nil {
			t.Fatal(err)
		}
	}
	<-done
	flush()
	saved, err := model.GetDeviceByDevID(original.ID)
	if err != nil {
		t.Fatal(err)
	}
	expected := uint64(30 * packed.Len())
	if saved.RX != expected || saved.TX != expected || !saved.Online {
		t.Fatalf("traffic/online mismatch: %+v; expected %d", saved, expected)
	}
	second := connect()
	replacement := waitDevice(t, network.ID, func(d model.Device) bool { return d.Online && d.SessionID != original.SessionID })
	if replacement.RX != expected || replacement.TX != expected {
		t.Fatalf("reconnect lost traffic: %+v", replacement)
	}
	_ = first.Close()
	flush()
	fresh, err := model.GetDeviceByDevID(original.ID)
	if err != nil || !fresh.Online || fresh.SessionID != replacement.SessionID {
		t.Fatalf("old disconnect replaced new session: %+v %v", fresh, err)
	}
	_ = second.Close()
	waitDevice(t, network.ID, func(d model.Device) bool { return !d.Online })
}
func TestFailedOfflineWriteIsRetriedAndStaleSessionIsDiscarded(t *testing.T) {
	_, network := runtimeAccount(t, "offlinetest")
	device := model.Device{NetID: network.ID, VMac: "offline", SessionID: "offline-old", IP: "192.168.202.2"}
	if err := device.Activate(); err != nil {
		t.Fatal(err)
	}
	db := storage.Get()
	statement := fmt.Sprintf("CREATE TRIGGER reject_offline BEFORE UPDATE OF online ON devices WHEN OLD.id = %d AND NEW.online = 0 BEGIN SELECT RAISE(ABORT, 'injected offline failure'); END", device.ID)
	if err := db.Exec(statement).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_offline") })
	device.Online = false
	persistDisconnect(device)
	offlineWrites.Lock()
	_, queued := offlineWrites.devices[device.SessionID]
	offlineWrites.Unlock()
	if !queued {
		t.Fatal("offline failure was lost")
	}
	if err := db.Exec("DROP TRIGGER reject_offline").Error; err != nil {
		t.Fatal(err)
	}
	retryDisconnects(context.Background())
	saved, err := model.GetDeviceByDevID(device.ID)
	if err != nil || saved.Online {
		t.Fatalf("offline retry failed: %+v %v", saved, err)
	}
	next := model.Device{NetID: network.ID, VMac: device.VMac, SessionID: "offline-new", IP: device.IP}
	if err := next.Activate(); err != nil {
		t.Fatal(err)
	}
	offlineWrites.Lock()
	offlineWrites.devices[device.SessionID] = device
	offlineWrites.Unlock()
	retryDisconnects(context.Background())
	saved, err = model.GetDeviceByDevID(device.ID)
	if err != nil || !saved.Online || saved.SessionID != next.SessionID {
		t.Fatalf("stale retry changed live session: %+v %v", saved, err)
	}
}
func TestReplacedRuntimeRejectsPendingConnections(t *testing.T) {
	_, network := runtimeAccount(t, "retiredtest")
	old := getNetById(network.ID)
	clone := *network
	UpdateNet(&clone)
	old.stateMu.RLock()
	retired := old.closed
	old.stateMu.RUnlock()
	if !retired || getNetById(network.ID) == old {
		t.Fatal("old network remained usable")
	}
	ReloadNet(network.ID)
	current := getNetById(network.ID)
	current.stateMu.RLock()
	retired = current.closed
	current.stateMu.RUnlock()
	if retired {
		t.Fatal("route reload retired the active network")
	}
}
