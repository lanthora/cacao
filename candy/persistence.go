package candy

import (
	"context"
	"errors"
	"sync"

	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"gorm.io/gorm"
)

var offlineWrites = struct {
	sync.Mutex
	devices map[string]model.Device
}{devices: map[string]model.Device{}}

func persistDisconnect(device model.Device) {
	if err := device.SaveRxTxOnline(); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Warn("persist disconnected device %d (queued for retry): %v", device.ID, err)
		offlineWrites.Lock()
		offlineWrites.devices[device.SessionID] = device
		offlineWrites.Unlock()
	}
}
func retryDisconnects(ctx context.Context) {
	offlineWrites.Lock()
	snapshot := make(map[string]model.Device, len(offlineWrites.devices))
	for key, d := range offlineWrites.devices {
		snapshot[key] = d
	}
	offlineWrites.Unlock()
	for key, d := range snapshot {
		err := d.SaveRxTxOnline(ctx)
		if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
			offlineWrites.Lock()
			delete(offlineWrites.devices, key)
			offlineWrites.Unlock()
		} else {
			logger.Warn("retry disconnected device %d: %v", d.ID, err)
		}
	}
}
