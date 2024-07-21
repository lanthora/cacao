package candy

const (
	AUTH      uint8 = 0
	FORWARD   uint8 = 1
	DHCP      uint8 = 2
	PEER      uint8 = 3
	VMAC      uint8 = 4
	DISCOVERY uint8 = 5
	ROUTE     uint8 = 6
	GENERAL   uint8 = 255
)

type AuthMessage struct {
	Type      uint8    `struc:"uint8"`
	IP        uint32   `struc:"uint32"`
	Timestamp int64    `struc:"int64"`
	Hash      [32]byte `struc:"[32]byte"`
}

type ForwardMessage struct {
	Type   uint8    `struc:"uint8"`
	Unused [12]byte `struc:"[12]byte"`
	Src    uint32   `struc:"uint32"`
	Dst    uint32   `struc:"uint32"`
}

type DHCPMessage struct {
	Type      uint8    `struc:"uint8"`
	Timestamp int64    `struc:"int64"`
	Cidr      []byte   `struc:"[32]byte"`
	Hash      [32]byte `struc:"[32]byte"`
}

type PeerConnMessage struct {
	Type uint8  `struc:"uint8"`
	Src  uint32 `struc:"uint32"`
	Dst  uint32 `struc:"uint32"`
	IP   uint32 `struc:"uint32"`
	Port uint16 `struc:"uint16"`
}

type VMacMessage struct {
	Type      uint8    `struc:"uint8"`
	VMac      string   `struc:"[16]byte"`
	Timestamp int64    `struc:"int64"`
	Hash      [32]byte `struc:"[32]byte"`
}

type DiscoveryMessage struct {
	Type uint8  `struc:"uint8"`
	Src  uint32 `struc:"uint32"`
	Dst  uint32 `struc:"uint32"`
}

type RouteMessage struct {
	Type     uint8  `struc:"uint8"`
	Size     uint8  `struc:"uint8"`
	Reserved uint16 `struc:"uint16"`
}

type RouteMessageEntry struct {
	Dest    uint32 `struc:"uint32"`
	Mask    uint32 `struc:"uint32"`
	NextHop uint32 `struc:"uint32"`
}

type GeneralMessage struct {
	Type    uint8  `struc:"uint8"`
	Subtype uint8  `struc:"uint8"`
	Extra   uint16 `struc:"uint16"`
	Src     uint32 `struc:"uint32"`
	Dst     uint32 `struc:"uint32"`
}
