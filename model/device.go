package model

import (
	"context"
	"sync"
	"time"

	"gorm.io/gorm"
)

type Device struct {
	// Keep atomic traffic counters aligned on 32-bit ARM as well.
	RX uint64
	TX uint64
	gorm.Model
	NetID     uint
	VMac      string
	SessionID string `gorm:"-" json:"-"`
	IP        string
	Online    bool
	OS        string
	Version   string
	Hostname  string
	Country   string
	Region    string
}

func ResetOnlineDevices(contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error {
		return tx.Model(&Device{}).Where("online = true").UpdateColumn("online", false).Error
	})
}

// Sessions exist only in this process; the original SQLite schema is unchanged.
// Each identity's operation lock orders activation and delayed session writes.
// Acquire it before entering SQLite; transaction callbacks never acquire it.
type deviceKey struct {
	network uint
	vmac    string
}
type deviceSession struct {
	sync.Mutex
	token string
}

var deviceSessions sync.Map

func sessionFor(d *Device) *deviceSession {
	value, _ := deviceSessions.LoadOrStore(deviceKey{d.NetID, d.VMac}, &deviceSession{})
	return value.(*deviceSession)
}
func attachSession(d *Device) {
	if value, ok := deviceSessions.Load(deviceKey{d.NetID, d.VMac}); ok {
		state := value.(*deviceSession)
		state.Lock()
		d.SessionID = state.token
		state.Unlock()
	}
}
func sessionWrite(d *Device, contexts []context.Context, fn func(*gorm.DB) error) error {
	if d.ID == 0 || d.SessionID == "" {
		return gorm.ErrRecordNotFound
	}
	value, ok := deviceSessions.Load(deviceKey{d.NetID, d.VMac})
	if !ok {
		return gorm.ErrRecordNotFound
	}
	state := value.(*deviceSession)
	state.Lock()
	defer state.Unlock()
	if state.token != d.SessionID {
		return gorm.ErrRecordNotFound
	}
	return writeDB(contexts, fn)
}

// Activate preserves persisted traffic and publishes the in-memory session only
// after commit. This requires one Cacao process per storage directory.
func (d *Device) Activate(contexts ...context.Context) error {
	input := *d
	state := sessionFor(d)
	state.Lock()
	defer state.Unlock()
	var saved Device
	err := writeDB(contexts, func(tx *gorm.DB) error {
		var parent Net
		if err := tx.First(&parent, input.NetID).Error; err != nil {
			return err
		}
		saved = Device{}
		var rows []Device
		if err := tx.Where("net_id = ? AND vmac = ?", input.NetID, input.VMac).Limit(2).Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) > 1 {
			return ErrAmbiguousData
		}
		if len(rows) == 1 {
			saved = rows[0]
		}
		saved.NetID = input.NetID
		saved.VMac = input.VMac
		saved.SessionID = input.SessionID
		saved.IP = input.IP
		saved.Online = true
		saved.Country = input.Country
		saved.Region = input.Region
		if err := tx.Model(&Device{}).Where("net_id = ? AND ip = ? AND online = true", input.NetID, input.IP).Update("online", false).Error; err != nil {
			return err
		}
		if saved.ID == 0 {
			return tx.Create(&saved).Error
		}
		return changed(tx.Model(&Device{}).Where("id = ?", saved.ID).Updates(map[string]interface{}{"ip": saved.IP, "online": true, "country": saved.Country, "region": saved.Region}))
	})
	if err == nil {
		state.token = input.SessionID
		*d = saved
	}
	return err
}
func (d *Device) SaveLocation(contexts ...context.Context) error {
	return sessionWrite(d, contexts, func(tx *gorm.DB) error {
		return changed(tx.Model(&Device{}).Where("id = ? AND net_id = ? AND vmac = ?", d.ID, d.NetID, d.VMac).Updates(map[string]interface{}{"country": d.Country, "region": d.Region}))
	})
}
func (d *Device) SaveTraffic(contexts ...context.Context) error {
	return sessionWrite(d, contexts, func(tx *gorm.DB) error {
		return changed(tx.Model(&Device{}).Where("id = ? AND net_id = ? AND vmac = ?", d.ID, d.NetID, d.VMac).Updates(map[string]interface{}{"rx": gorm.Expr("max(rx, ?)", d.RX), "tx": gorm.Expr("max(tx, ?)", d.TX)}))
	})
}
func (d *Device) SaveRxTxOnline(contexts ...context.Context) error {
	if d.ID == 0 {
		return gorm.ErrRecordNotFound
	}
	return sessionWrite(d, contexts, func(tx *gorm.DB) error {
		return changed(tx.Model(&Device{}).Where("id = ? AND net_id = ? AND vmac = ?", d.ID, d.NetID, d.VMac).Updates(map[string]interface{}{"rx": gorm.Expr("max(rx, ?)", d.RX), "tx": gorm.Expr("max(tx, ?)", d.TX), "online": d.Online}))
	})
}
func (d *Device) SaveOsVersionHostname(contexts ...context.Context) error {
	if d.ID == 0 {
		return gorm.ErrRecordNotFound
	}
	return sessionWrite(d, contexts, func(tx *gorm.DB) error {
		return changed(tx.Model(&Device{}).Where("id = ? AND net_id = ? AND vmac = ?", d.ID, d.NetID, d.VMac).Updates(map[string]interface{}{"os": d.OS, "version": d.Version, "hostname": d.Hostname}))
	})
}
func (d *Device) Delete(contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error { return changed(tx.Where("id = ? AND online = false", d.ID).Delete(&Device{})) })
}
func GetDeviceByDevID(id uint, contexts ...context.Context) (Device, error) {
	var row Device
	if id == 0 {
		return row, gorm.ErrRecordNotFound
	}
	err := readDB(contexts).First(&row, id).Error
	if err == nil {
		attachSession(&row)
	}
	return row, err
}
func GetDevicesByNetID(id uint, contexts ...context.Context) ([]Device, error) {
	var rows []Device
	err := readDB(contexts).Where("net_id = ?", id).Find(&rows).Error
	if err == nil {
		for i := range rows {
			attachSession(&rows[i])
		}
	}
	return rows, err
}
func GetDevicesByUserID(id uint, contexts ...context.Context) ([]Device, error) {
	var rows []Device
	err := readDB(contexts).Joins("JOIN nets ON nets.id = devices.net_id AND nets.deleted_at IS NULL").Where("nets.user_id = ?", id).Find(&rows).Error
	if err == nil {
		for i := range rows {
			attachSession(&rows[i])
		}
	}
	return rows, err
}
func GetTrafficByUserID(id uint, contexts ...context.Context) (uint64, uint64, error) {
	var sums struct {
		RX uint64
		TX uint64
	}
	err := readDB(contexts).Unscoped().Model(&Device{}).Select("coalesce(sum(rx),0) AS rx, coalesce(sum(tx),0) AS tx").Joins("JOIN nets ON nets.id = devices.net_id").Where("nets.user_id = ?", id).Scan(&sums).Error
	return sums.RX, sums.TX, err
}
func DeleteExpiredDevice(id uint, cutoff time.Time, contexts ...context.Context) error {
	return writeDB(contexts, func(tx *gorm.DB) error {
		return tx.Where("id = ? AND online = false AND updated_at < ?", id, cutoff).Delete(&Device{}).Error
	})
}
