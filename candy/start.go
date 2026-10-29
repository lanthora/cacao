package candy

import (
	"context"
	"time"

	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
)

// Start runs only after existing-schema validation (or initial table creation). Tests import candy without opening a
// database, mutating online flags or starting unmanaged background goroutines.
func Start(ctx context.Context) error {
	if err := model.ResetOnlineDevices(ctx); err != nil {
		return err
	}
	nets, err := model.GetNets(ctx)
	if err != nil {
		return err
	}
	for i := range nets {
		InsertNet(&nets[i])
	}
	go periodic(ctx, time.Minute, func() { flush(ctx) })
	go periodic(ctx, time.Hour, func() {
		enabled, err := model.GetConfig("autoCleanUser", "false", ctx)
		if err == nil && enabled == "true" {
			err = CleanInactiveUser(ctx)
		}
		if err != nil {
			logger.Warn("clean inactive users: %v", err)
		}
	})
	go periodic(ctx, time.Hour, func() {
		if err := cleanInactiveDevices(ctx); err != nil {
			logger.Warn("clean inactive devices: %v", err)
		}
	})
	return nil
}
func periodic(ctx context.Context, interval time.Duration, fn func()) {
	if ctx.Err() != nil {
		return
	}
	fn()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}
