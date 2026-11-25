package candy

import (
	"crypto/rand"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lanthora/cacao/logger"
)

const (
	relayQueueSize    = 1024
	forwardMessageLen = 21
)

var (
	// Overridable in tests. A slot that has been silent for longer than this
	// can be replaced before the data connection read timeout fires.
	relayStaleAfter      = 25 * time.Second
	relayPingInterval    = 15 * time.Second
	relayDataReadTimeout = 45 * time.Second
)

type relaySlot struct {
	ws         *candysocket
	lastActive int64 // UnixNano, guarded by relay.mu
}

// relay holds the negotiated data connections of one control connection.
// Incoming packets for the device are queued here and distributed over the
// live data connections; when no data connection is alive the caller falls
// back to the control connection.
type relay struct {
	key        [32]byte
	target     int
	staleAfter time.Duration

	mu     sync.Mutex
	closed bool
	slots  map[*candysocket]*relaySlot

	live  atomic.Int32
	queue chan []byte
	done  chan struct{}
}

func newRelay(key [32]byte, target int) *relay {
	return &relay{
		key:        key,
		target:     target,
		staleAfter: relayStaleAfter,
		slots:      make(map[*candysocket]*relaySlot),
		queue:      make(chan []byte, relayQueueSize),
		done:       make(chan struct{}),
	}
}

func (r *relay) matches(key [32]byte) bool {
	return r.key == key
}

// admit adds a data connection. While at capacity it replaces a slot that has
// not delivered any frame for staleAfter; otherwise it allows one grace slot so
// a fast peer-side reconnect is not blocked by a connection the server has not
// timed out yet.
func (r *relay) admit(ws *candysocket, now int64) bool {
	var evicted *candysocket
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return false
	}
	if len(r.slots) >= r.target {
		var stale *relaySlot
		for _, slot := range r.slots {
			if now-slot.lastActive > int64(r.staleAfter) && (stale == nil || slot.lastActive < stale.lastActive) {
				stale = slot
			}
		}
		if stale != nil {
			delete(r.slots, stale.ws)
			r.live.Add(-1)
			evicted = stale.ws
		} else if len(r.slots) >= r.target+1 {
			r.mu.Unlock()
			return false
		}
	}
	r.slots[ws] = &relaySlot{ws: ws, lastActive: now}
	r.live.Add(1)
	r.mu.Unlock()
	if evicted != nil {
		logger.Debug("relay evicted stale data connection: target=%d", r.target)
		evicted.conn.Close()
	}
	return true
}

func (r *relay) touch(ws *candysocket, now int64) {
	r.mu.Lock()
	if slot, ok := r.slots[ws]; ok {
		slot.lastActive = now
	}
	r.mu.Unlock()
}

func (r *relay) unregister(ws *candysocket) {
	drained := false
	r.mu.Lock()
	if _, ok := r.slots[ws]; ok {
		delete(r.slots, ws)
		if r.live.Add(-1) == 0 {
			drained = true
		}
	}
	r.mu.Unlock()
	if drained {
		r.drain()
	}
}

func (r *relay) enqueue(buffer []byte) bool {
	if r.live.Load() == 0 {
		return false
	}
	select {
	case <-r.done:
		return false
	default:
	}
	select {
	case r.queue <- buffer:
	default:
		// Mirror the single-connection relay: drop when the queue is full
		// instead of mixing data paths and reordering packets.
	}
	return true
}

func (r *relay) stop() {
	conns := make([]*candysocket, 0)
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	close(r.done)
	for _, slot := range r.slots {
		conns = append(conns, slot.ws)
	}
	r.slots = make(map[*candysocket]*relaySlot)
	r.live.Store(0)
	r.mu.Unlock()
	for _, ws := range conns {
		ws.conn.Close()
	}
	r.drain()
}

func (r *relay) drain() {
	for {
		select {
		case <-r.queue:
		default:
			return
		}
	}
}

func (r *relay) writeLoop(ws *candysocket, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case <-r.done:
			return
		case buffer := <-r.queue:
			if err := ws.writeMessage(buffer); err != nil {
				ws.conn.Close()
				return
			}
			if ws.dev != nil {
				atomic.AddUint64(&ws.dev.model.RX, uint64(len(buffer)))
			}
		}
	}
}

func (r *relay) pingLoop(ws *candysocket, done <-chan struct{}) {
	ticker := time.NewTicker(relayPingInterval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-r.done:
			return
		case <-ticker.C:
			if err := ws.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second)); err != nil {
				ws.conn.Close()
				return
			}
		}
	}
}

// serveRelay owns the read loop of one accepted data connection. The connection
// is already registered by admit(); it is unregistered when the loop exits.
func (ws *candysocket) serveRelay() {
	r := ws.relay
	if r == nil {
		return
	}
	done := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(2)
	go func() {
		defer workers.Done()
		r.writeLoop(ws, done)
	}()
	go func() {
		defer workers.Done()
		r.pingLoop(ws, done)
	}()
	defer func() {
		close(done)
		ws.conn.Close()
		workers.Wait()
		r.unregister(ws)
	}()

	ws.conn.SetPingHandler(func(buffer string) error {
		ws.conn.SetReadDeadline(time.Now().Add(relayDataReadTimeout))
		r.touch(ws, time.Now().UnixNano())
		return ws.writePong([]byte(buffer))
	})
	ws.conn.SetPongHandler(func(string) error {
		ws.conn.SetReadDeadline(time.Now().Add(relayDataReadTimeout))
		r.touch(ws, time.Now().UnixNano())
		return nil
	})

	for {
		ws.conn.SetReadDeadline(time.Now().Add(relayDataReadTimeout))
		messageType, buffer, err := ws.conn.ReadMessage()
		if err != nil {
			return
		}
		r.touch(ws, time.Now().UnixNano())
		if messageType != websocket.BinaryMessage {
			return
		}
		if len(buffer) < forwardMessageLen || buffer[0] != FORWARD {
			return
		}
		if err := ws.handleRelayForward(buffer); err != nil {
			return
		}
	}
}

func (ws *candysocket) handleRelayForward(buffer []byte) error {
	ws.net.stateMu.RLock()
	defer ws.net.stateMu.RUnlock()
	if ws.net.closed {
		return fmt.Errorf("network was removed or replaced")
	}
	return ws.handleForwardMessage(buffer)
}

// sendPacket delivers a forwarded packet to the device: through the relay queue
// while a data connection is alive, otherwise directly on the control connection.
func (ws *candysocket) sendPacket(buffer []byte) {
	if ws.relay != nil && ws.relay.enqueue(buffer) {
		return
	}
	if err := ws.writeMessage(buffer); err == nil && ws.dev != nil {
		atomic.AddUint64(&ws.dev.model.RX, uint64(len(buffer)))
	}
}

func (ws *candysocket) stopRelay() {
	if ws.relay != nil {
		ws.relay.stop()
	}
}

func (ws *candysocket) handleMultiMessage(buffer []byte) error {
	if len(buffer) != 34 {
		return fmt.Errorf("relay negotiation failed: invalid length %d", len(buffer))
	}
	if ws.dev == nil || !ws.dev.model.Online || ws.relay != nil {
		return nil
	}
	count := buffer[1]
	if count == 0 {
		return nil
	}
	if count > 7 {
		count = 7
	}
	var key [32]byte
	if _, err := rand.Read(key[:]); err != nil {
		logger.Warn("relay key generation failed: %v", err)
		return nil
	}
	echo := make([]byte, len(buffer))
	copy(echo, buffer)
	echo[1] = count
	copy(echo[2:], key[:])
	ws.relay = newRelay(key, int(count))
	if err := ws.writeMessage(echo); err != nil {
		ws.relay = nil
		return err
	}
	logger.Debug("relay negotiated: target=%d", count)
	return nil
}

func (ws *candysocket) handleJoinMessage(buffer []byte) error {
	if len(buffer) != 34 {
		return fmt.Errorf("relay join failed: invalid length %d", len(buffer))
	}
	if ws.dev != nil && ws.dev.model.Online {
		return fmt.Errorf("relay join failed: connection is already authorized")
	}
	var key [32]byte
	copy(key[:], buffer[2:34])

	var control *candysocket
	ws.net.ipWsMapMutex.RLock()
	for _, candidate := range ws.net.ipWsMap {
		if candidate.relay == nil || candidate.dev == nil || !candidate.dev.model.Online {
			continue
		}
		if candidate.relay.matches(key) {
			control = candidate
			break
		}
	}
	ws.net.ipWsMapMutex.RUnlock()
	if control == nil {
		return fmt.Errorf("relay join failed: unknown key")
	}
	if !control.relay.admit(ws, time.Now().UnixNano()) {
		return fmt.Errorf("relay join failed: no capacity")
	}
	ws.dev = control.dev
	ws.relay = control.relay
	ws.data = true
	if err := ws.writeMessage(buffer); err != nil {
		control.relay.unregister(ws)
		return err
	}
	logger.Debug("relay data connection joined: target=%d", control.relay.target)
	return nil
}
