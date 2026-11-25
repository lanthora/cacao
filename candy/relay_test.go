package candy

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lanthora/cacao/model"
	"github.com/lunixbochs/struc"
)

type relayEnv struct {
	t    *testing.T
	user model.User
	net  *model.Net
	url  string
}

func newRelayEnv(t *testing.T, name string) *relayEnv {
	t.Helper()
	user, network := runtimeAccount(t, name)
	router := gin.New()
	router.Use(WebsocketMiddleware())
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return &relayEnv{
		t:    t,
		user: user,
		net:  network,
		url:  "ws" + strings.TrimPrefix(server.URL, "http"),
	}
}

func (e *relayEnv) dial() *websocket.Conn {
	e.t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(e.url+"/"+e.user.Name, http.Header{"X-Forwarded-For": []string{"10.0.0.1"}})
	if err != nil {
		e.t.Fatal(err)
	}
	e.t.Cleanup(func() { conn.Close() })
	return conn
}

func (e *relayEnv) login(vmac, ip string) *websocket.Conn {
	e.t.Helper()
	conn := e.dial()
	vmacMessage := VMacMessage{Type: VMAC, VMac: vmac, Timestamp: time.Now().Unix()}
	payload := append([]byte(e.net.Password), []byte(vmac)...)
	payload = binary.BigEndian.AppendUint64(payload, uint64(vmacMessage.Timestamp))
	vmacMessage.Hash = sha256.Sum256(payload)
	relaySend(e.t, conn, &vmacMessage)
	authMessage := AuthMessage{Type: AUTH, IP: strIpToUint32(ip), Timestamp: time.Now().Unix()}
	payload = binary.BigEndian.AppendUint32([]byte(e.net.Password), authMessage.IP)
	payload = binary.BigEndian.AppendUint64(payload, uint64(authMessage.Timestamp))
	authMessage.Hash = sha256.Sum256(payload)
	relaySend(e.t, conn, &authMessage)
	waitDevice(e.t, e.net.ID, func(d model.Device) bool { return d.IP == ip && d.Online })
	return conn
}

func relayPack(t *testing.T, message interface{}) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := struc.Pack(&buffer, message); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func relaySend(t *testing.T, conn *websocket.Conn, message interface{}) {
	t.Helper()
	if err := conn.WriteMessage(websocket.BinaryMessage, relayPack(t, message)); err != nil {
		t.Fatal(err)
	}
}

func relaySendRaw(t *testing.T, conn *websocket.Conn, buffer []byte) {
	t.Helper()
	if err := conn.WriteMessage(websocket.BinaryMessage, buffer); err != nil {
		t.Fatal(err)
	}
}

func relayRead(t *testing.T, conn *websocket.Conn, timeout time.Duration) ([]byte, error) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(timeout))
	_, buffer, err := conn.ReadMessage()
	return buffer, err
}

func relayNegotiate(t *testing.T, conn *websocket.Conn, count uint8) [32]byte {
	t.Helper()
	relaySend(t, conn, &MultiMessage{Type: MULTI, Count: count})
	buffer, err := relayRead(t, conn, 3*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(buffer) != 34 {
		t.Fatalf("invalid MULTI echo length: %d", len(buffer))
	}
	var echo MultiMessage
	if err := struc.Unpack(bytes.NewReader(buffer), &echo); err != nil {
		t.Fatal(err)
	}
	if echo.Type != MULTI || echo.Count != count {
		t.Fatalf("invalid MULTI echo: type=%d count=%d", echo.Type, echo.Count)
	}
	return echo.Key
}

func relayJoin(t *testing.T, env *relayEnv, key [32]byte, count uint8) (*websocket.Conn, error) {
	t.Helper()
	message := MultiMessage{Type: JOIN, Count: count, Key: key}
	conn := env.dial()
	relaySend(t, conn, &message)
	buffer, err := relayRead(t, conn, 3*time.Second)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if !bytes.Equal(buffer, relayPack(t, &message)) {
		conn.Close()
		t.Fatalf("invalid JOIN echo: %x", buffer)
	}
	return conn, nil
}

func relayJoinRetry(t *testing.T, env *relayEnv, key [32]byte, count uint8, timeout time.Duration) *websocket.Conn {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		conn, err := relayJoin(t, env, key, count)
		if err == nil {
			return conn
		}
		if time.Now().After(deadline) {
			t.Fatalf("relay join did not succeed: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

type frameReader struct {
	mu    sync.Mutex
	count int
	done  chan struct{}
}

func startFrameReader(conn *websocket.Conn) *frameReader {
	reader := &frameReader{done: make(chan struct{})}
	go func() {
		defer close(reader.done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
			reader.mu.Lock()
			reader.count++
			reader.mu.Unlock()
		}
	}()
	return reader
}

func (r *frameReader) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}

func (r *frameReader) Done() <-chan struct{} {
	return r.done
}

func waitFrameCount(t *testing.T, want int, timeout time.Duration, readers ...*frameReader) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		total := 0
		for _, reader := range readers {
			total += reader.Count()
		}
		if total >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	total := 0
	for _, reader := range readers {
		total += reader.Count()
	}
	t.Fatalf("expected %d frames, got %d", want, total)
}

func waitReaderClosed(t *testing.T, reader *frameReader, timeout time.Duration) {
	t.Helper()
	select {
	case <-reader.Done():
	case <-time.After(timeout):
		t.Fatal("data connection was not closed")
	}
}

func TestRelayNegotiationAdmissionAndForwarding(t *testing.T) {
	env := newRelayEnv(t, "relayforward")
	a := env.login("0000000000000011", "192.168.202.10")
	b := env.login("0000000000000012", "192.168.202.11")
	key := relayNegotiate(t, a, 2)

	d1, err := relayJoin(t, env, key, 2)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := relayJoin(t, env, key, 2)
	if err != nil {
		t.Fatal(err)
	}
	d3, err := relayJoin(t, env, key, 2)
	if err != nil {
		t.Fatalf("grace slot join failed: %v", err)
	}

	rejected := env.dial()
	relaySend(t, rejected, &MultiMessage{Type: JOIN, Count: 2, Key: key})
	if _, err := relayRead(t, rejected, time.Second); err == nil {
		t.Fatal("join beyond the grace slot was accepted")
	}
	rejected.Close()

	wrong := env.dial()
	relaySend(t, wrong, &MultiMessage{Type: JOIN, Count: 2, Key: [32]byte{}})
	if _, err := relayRead(t, wrong, time.Second); err == nil {
		t.Fatal("join with a wrong key was accepted")
	}
	wrong.Close()

	time.Sleep(200 * time.Millisecond)
	dataReaders := []*frameReader{startFrameReader(d1), startFrameReader(d2), startFrameReader(d3)}
	controlReader := startFrameReader(a)

	src := strIpToUint32("192.168.202.11")
	dst := strIpToUint32("192.168.202.10")
	for i := 0; i < 30; i++ {
		relaySend(t, b, &ForwardMessage{Type: FORWARD, Src: src, Dst: dst})
		time.Sleep(2 * time.Millisecond)
	}
	waitFrameCount(t, 30, 5*time.Second, dataReaders...)
	time.Sleep(100 * time.Millisecond)
	if got := controlReader.Count(); got != 0 {
		t.Fatalf("control connection received %d forwarded packets", got)
	}
}

func TestRelayReplenishmentAfterLoss(t *testing.T) {
	env := newRelayEnv(t, "relayreplenish")
	a := env.login("0000000000000021", "192.168.202.30")
	b := env.login("0000000000000022", "192.168.202.31")
	key := relayNegotiate(t, a, 1)

	first, err := relayJoin(t, env, key, 1)
	if err != nil {
		t.Fatal(err)
	}
	first.Close()

	replacement := relayJoinRetry(t, env, key, 1, 5*time.Second)
	dataReader := startFrameReader(replacement)
	controlReader := startFrameReader(a)

	src := strIpToUint32("192.168.202.31")
	dst := strIpToUint32("192.168.202.30")
	for i := 0; i < 5; i++ {
		relaySend(t, b, &ForwardMessage{Type: FORWARD, Src: src, Dst: dst})
	}
	waitFrameCount(t, 5, 5*time.Second, dataReader)
	time.Sleep(100 * time.Millisecond)
	if got := controlReader.Count(); got != 0 {
		t.Fatalf("control connection received %d forwarded packets", got)
	}
}

func TestRelayStaleEvictionAndControlClose(t *testing.T) {
	previous := relayStaleAfter
	relayStaleAfter = 300 * time.Millisecond
	t.Cleanup(func() { relayStaleAfter = previous })

	env := newRelayEnv(t, "relaystale")
	a := env.login("0000000000000031", "192.168.202.40")
	key := relayNegotiate(t, a, 1)
	first, err := relayJoin(t, env, key, 1)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(600 * time.Millisecond)
	second, err := relayJoin(t, env, key, 1)
	if err != nil {
		t.Fatalf("stale slot was not replaced: %v", err)
	}
	if _, err := relayRead(t, first, 2*time.Second); err == nil {
		t.Fatal("stale data connection was not closed")
	}

	a.Close()
	if _, err := relayRead(t, second, 2*time.Second); err == nil {
		t.Fatal("closing the control connection did not stop the relay")
	}
	waitDevice(t, env.net.ID, func(d model.Device) bool { return d.IP == "192.168.202.40" && !d.Online })
}

func TestRelayFallbackAndDataConnectionValidation(t *testing.T) {
	env := newRelayEnv(t, "relayfallback")
	a := env.login("0000000000000041", "192.168.202.50")
	b := env.login("0000000000000042", "192.168.202.51")
	key := relayNegotiate(t, a, 2)

	controlReader := startFrameReader(a)
	src := strIpToUint32("192.168.202.51")
	dst := strIpToUint32("192.168.202.50")
	relaySend(t, b, &ForwardMessage{Type: FORWARD, Src: src, Dst: dst})
	waitFrameCount(t, 1, 3*time.Second, controlReader)

	data, err := relayJoin(t, env, key, 2)
	if err != nil {
		t.Fatal(err)
	}
	dataReader := startFrameReader(data)
	relaySend(t, b, &ForwardMessage{Type: FORWARD, Src: src, Dst: dst})
	waitFrameCount(t, 1, 3*time.Second, dataReader)

	relaySendRaw(t, data, []byte{DHCP})
	waitReaderClosed(t, dataReader, 3*time.Second)
}

func TestRelayBroadcastExcludesSender(t *testing.T) {
	env := newRelayEnv(t, "relaybroadcast")
	a := env.login("0000000000000051", "192.168.202.60")
	b := env.login("0000000000000052", "192.168.202.61")
	key := relayNegotiate(t, a, 1)
	data, err := relayJoin(t, env, key, 1)
	if err != nil {
		t.Fatal(err)
	}

	aReader := startFrameReader(a)
	bReader := startFrameReader(b)
	relaySend(t, data, &ForwardMessage{Type: FORWARD, Src: strIpToUint32("192.168.202.60"), Dst: 0xFFFFFFFF})
	waitFrameCount(t, 1, 3*time.Second, bReader)
	time.Sleep(200 * time.Millisecond)
	if got := aReader.Count(); got != 0 {
		t.Fatalf("sender received its own broadcast %d times", got)
	}
}

func TestRelayReplacedControlClosesDataConnection(t *testing.T) {
	env := newRelayEnv(t, "relayreplace")
	a := env.login("0000000000000061", "192.168.202.70")
	key := relayNegotiate(t, a, 1)
	data, err := relayJoin(t, env, key, 1)
	if err != nil {
		t.Fatal(err)
	}

	env.login("0000000000000061", "192.168.202.70")
	waitDevice(t, env.net.ID, func(d model.Device) bool { return d.IP == "192.168.202.70" && d.Online })
	if _, err := relayRead(t, data, 2*time.Second); err == nil {
		t.Fatal("data connection survived control connection replacement")
	}
}

func TestRelayDeviceIPUpdateClosesDataConnection(t *testing.T) {
	env := newRelayEnv(t, "relayipupdate")
	a := env.login("0000000000000071", "192.168.202.80")
	key := relayNegotiate(t, a, 1)
	data, err := relayJoin(t, env, key, 1)
	if err != nil {
		t.Fatal(err)
	}
	device := waitDevice(t, env.net.ID, func(d model.Device) bool { return d.IP == "192.168.202.80" && d.Online })
	if _, err := UpdateDeviceIP(context.Background(), env.user.ID, device.ID, "192.168.202.81"); err != nil {
		t.Fatal(err)
	}
	if _, err := relayRead(t, data, 2*time.Second); err == nil {
		t.Fatal("data connection survived the device IP update")
	}
}
