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

// snapshotDevice copies a live device without reading the atomic traffic
// counters non-atomically.
func snapshotDevice(device *model.Device) model.Device {
	return model.Device{
		Model:     device.Model,
		NetID:     device.NetID,
		VMac:      device.VMac,
		SessionID: device.SessionID,
		IP:        device.IP,
		Online:    device.Online,
		OS:        device.OS,
		Version:   device.Version,
		Hostname:  device.Hostname,
		Country:   device.Country,
		Region:    device.Region,
		RX:        atomic.LoadUint64(&device.RX),
		TX:        atomic.LoadUint64(&device.TX),
	}
}

// assignDevice updates every field except the atomic traffic counters. The
// counters are updated by packet handling goroutines and must never be written
// as part of a structure copy.
func assignDevice(dst, src *model.Device) {
	dst.Model = src.Model
	dst.NetID = src.NetID
	dst.VMac = src.VMac
	dst.SessionID = src.SessionID
	dst.IP = src.IP
	dst.Online = src.Online
	dst.OS = src.OS
	dst.Version = src.Version
	dst.Hostname = src.Hostname
	dst.Country = src.Country
	dst.Region = src.Region
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
			ws.stopRelay()
			delete(n.ipWsMap, ip)
			_ = ws.writeCloseMessage("ip changed")
			_ = ws.conn.Close()
		}
	}
	return saved, nil
}
