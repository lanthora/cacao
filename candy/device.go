package candy

import (
	"context"
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
