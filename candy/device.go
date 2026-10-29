package candy

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/lanthora/cacao/model"
)

func cleanInactiveDevices(ctx context.Context) error {
	nets, err := model.GetNets(ctx)
	if err != nil {
		return err
	}
	for _, n := range nets {
		if n.Lease > 0 {
			devices, err := model.GetDevicesByNetID(n.ID, ctx)
			if err != nil {
				return err
			}
			cutoff := time.Now().AddDate(0, 0, -int(n.Lease))
			for _, d := range devices {
				if !d.Online && d.UpdatedAt.Before(cutoff) {
					if err := model.DeleteExpiredDevice(d.ID, cutoff, ctx); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

type Device struct {
	model *model.Device
	ip    uint32
}

// UpdateDeviceIP orders validation/commit and targeted disconnection against
// network replacement, DHCP offers and authentication. No socket is closed on failure.
func UpdateDeviceIP(ctx context.Context, userID, deviceID uint, address string) (model.Device, error) {
	ConfigMu.Lock()
	defer ConfigMu.Unlock()
	device, err := model.GetDeviceByDevID(deviceID, ctx)
	if err != nil {
		return model.Device{}, err
	}
	n := getNetById(device.NetID)
	var rx, tx uint64
	if n != nil {
		n.stateMu.Lock()
		defer n.stateMu.Unlock()
		for _, ws := range n.ipWsMap {
			if ws.dev.model.ID == deviceID {
				rx = atomic.LoadUint64(&ws.dev.model.RX)
				tx = atomic.LoadUint64(&ws.dev.model.TX)
			}
		}
	}
	saved, updated, err := model.UpdateDeviceIP(userID, deviceID, address, rx, tx, ctx)
	if err != nil || !updated || n == nil {
		return saved, err
	}
	if n.addressRevisions == nil {
		n.addressRevisions = make(map[string]uint64)
	}
	n.addressRevisions[saved.VMac]++
	n.ipWsMapMutex.Lock()
	defer n.ipWsMapMutex.Unlock()
	for ip, ws := range n.ipWsMap {
		if ws.dev.model.ID == deviceID {
			ws.dev.model.Online = false
			delete(n.ipWsMap, ip)
			_ = ws.writeCloseMessage("ip changed")
			_ = ws.conn.Close()
		}
	}
	return saved, nil
}
