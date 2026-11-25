# Multi-connection relay

Cacao supports the negotiated multi-connection relay used by candy clients
started with `connections > 1` (CLI `--connections`, config `connections = N`,
1-8 including the control connection). The feature is transparent to older
clients and does not change the database schema or the REST API.

## Wire protocol

Two message kinds are added on top of the existing candy protocol:

| Kind | Value | Payload |
| --- | --- | --- |
| `MULTI` | 7 | 34 bytes: `type:u8`, `count:u8`, `key:[32]byte` |
| `JOIN` | 8 | the same 34 bytes, with `type = JOIN` |

Flow:

1. After authentication the client sends `MULTI` with `count = connections - 1`
   on the control connection.
2. The server generates a random 32-byte key, clamps `count` to 7 and echoes the
   message with the key filled in. `count == 0` and a second `MULTI` on the same
   control connection are ignored.
3. The client opens `count` more WebSocket connections and sends `JOIN` with the
   key. The server echoes the message unchanged and binds the connection to the
   device session. A `JOIN` with an unknown key, on an already authorized
   connection, or without free capacity is rejected by closing the connection.
4. Data connections carry only `FORWARD` binary frames. Authentication, DHCP,
   VMAC, PEER, discovery, routes and general messages stay on the control
   connection.

Keys are scoped to one network: a `JOIN` can only match a relay of a control
connection in the same network (the WebSocket path selects the network). The key
is random per session, is never logged and is not bound to a source address.

## Server behavior

- The server keeps the negotiated count as `target` and tracks the live data
  connections of the device. A data connection is admitted while fewer than
  `target` connections are live.
- Every data connection records the time of its last received frame. A slot that
  has not delivered any frame for 25 seconds can be replaced, which covers
  connections the peer has already abandoned but the server has not timed out yet
  (the relay read timeout is 45 seconds).
- One temporary extra connection (`target + 1`) is allowed while both sides
  converge, so a client that noticed the failure first is not blocked waiting for
  the server-side timeout.
- The server sends a WebSocket PING every 15 seconds on every data connection and
  answers client PINGs with PONG. Both actions refresh the last-frame time of the
  slot.
- Packets forwarded to the device are queued (1024 packets) and distributed over
  the live data connections. Packets are dropped when the queue is full, matching
  the single-connection relay behavior. When no data connection is alive the
  packet is written to the control connection.
- Broadcast and multicast forwarding uses the same queue for every destination
  device, but never sends a packet back to the device that originated it, even
  when it was sent over a data connection.
- Traffic counters belong to the device session: bytes written on data
  connections are counted as RX for that device, and both directions are included
  in the normal `/api/device/show` counters and the periodic flush.
- Data connections are not devices. They never enter the online device map, never
  mark a device offline, and are closed when the control connection closes or is
  replaced, when the device virtual IP changes, or when the network is reloaded
  or deleted.

## Client expectations

Initial negotiation is compatible with candy clients that implement `MULTI` /
`JOIN` (v6.1.12 and later). Automatic re-establishment of a lost data connection
requires a client with the matching supervisor loop; without it the client simply
keeps using the control connection after its data connections are gone. Cacao
accepts re-`JOIN`s with the same key for the lifetime of the session, so a patched
client can replenish lost connections without restarting the session.

A client that re-joins against a server without replenishment support receives a
closed connection and should retry with backoff while continuing to use the
control connection.

## Limits

| Item | Value |
| --- | --- |
| Extra connections per device | up to 7 (`target`) |
| Temporary extra connection | 1 (`target + 1`) |
| Stale slot threshold | 25 seconds |
| Data connection read timeout | 45 seconds |
| Data connection ping interval | 15 seconds |
| Relay queue | 1024 packets, drop on overflow |

The per-device relay queue is allocated only when a client negotiates more than
one connection.

## Compatibility

| Client | Result |
| --- | --- |
| older client without `MULTI` | never negotiates; single control connection, unchanged |
| candy v6.1.12+ without replenishment patch | negotiates at authentication; data connections are not rebuilt after loss |
| candy with replenishment patch | full replenishment |
| any client against a server without this feature | negotiation is ignored or rejected; client falls back to the control connection |

## Verification

`candy/relay_test.go` covers negotiation, the grace slot and capacity rejection,
wrong keys, forwarding over data connections, replacement after a loss, stale slot
eviction, fallback to the control connection, broadcast sender exclusion, control
connection replacement and device IP changes. Run:

```bash
go test -race ./...
```
