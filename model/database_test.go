package model

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

func resetData(t *testing.T) {
	t.Helper()
	clear := func() {
		for _, table := range []string{"routes", "devices", "nets", "users", "configs"} {
			if err := storage.Get().Exec("DELETE FROM " + table).Error; err != nil {
				t.Errorf("clear %s: %v", table, err)
			}
		}
	}
	clear()
	t.Cleanup(clear)
}
func createNormal(t *testing.T, name string) (User, *Net) {
	t.Helper()
	u, n, err := RegisterUser(User{Name: name, Password: "old-password", Token: "old-token"}, Net{Name: "@", DHCP: "192.168.202.0/24", Broadcast: true}, true)
	if err != nil {
		t.Fatal(err)
	}
	return u, n
}
func TestConcurrentFirstAdministratorAndUniqueUsers(t *testing.T) {
	resetData(t)
	var wg sync.WaitGroup
	errs := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u, n, err := RegisterUser(User{Name: fmt.Sprintf("first-%d", i), IP: fmt.Sprint(i)}, Net{Name: "@"}, false)
			if err == nil && (u.Role != "admin" || n != nil) {
				err = fmt.Errorf("invalid first account: %+v %+v", u, n)
			}
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrRegistrationDisabled) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("first admins: %d", success)
	}
	users, err := GetUsers()
	if err != nil || len(users) != 1 {
		t.Fatalf("users: %v %v", users, err)
	}
	if open, err := GetConfig("openreg", "true"); err != nil || open != "false" {
		t.Fatalf("registration not disabled: %s %v", open, err)
	}
	errs = make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := RegisterUser(User{Name: "same-user"}, Net{Name: "@"}, true)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	success = 0
	for err := range errs {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrUsernameTaken) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("duplicate username accepted %d times", success)
	}
}
func TestRegistrationRollsBackAllSteps(t *testing.T) {
	resetData(t)
	db := storage.Get()
	if err := db.Exec("CREATE TRIGGER reject_network BEFORE INSERT ON nets BEGIN SELECT RAISE(ABORT, 'injected network failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_network") })
	u, n, err := RegisterUser(User{Name: "atomic-user"}, Net{Name: "@"}, true)
	if err == nil || u.ID != 0 || n != nil {
		t.Fatalf("partial result: %+v %+v %v", u, n, err)
	}
	if _, err := FindUser("atomic-user"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("partial user committed: %v", err)
	}
	if err := db.Exec("DROP TRIGGER reject_network").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TRIGGER reject_registration_switch BEFORE INSERT ON configs BEGIN SELECT RAISE(ABORT, 'injected config failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_registration_switch") })
	if _, _, err := RegisterUser(User{Name: "atomic-admin"}, Net{}, false); err == nil {
		t.Fatal("expected failure")
	}
	users, err := GetUsers()
	if err != nil || len(users) != 0 {
		t.Fatalf("partial admin committed: %+v %v", users, err)
	}
}
func TestCredentialsAreNotOverwrittenByActivityOrStaleRequests(t *testing.T) {
	resetData(t)
	user, _ := createNormal(t, "credentials")
	stale := user
	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- RefreshUserLastActiveTimeByUserID(user.ID) }()
	}
	if err := user.UpdateCredentials("new-password", "new-token", "new-ip"); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := stale.UpdateCredentials("old-password", "stale-token", ""); !errors.Is(err, ErrCredentialsChanged) {
		t.Fatalf("stale request: %v", err)
	}
	found, err := FindUser(user.Name)
	if err != nil {
		t.Fatal(err)
	}
	if found.Password != "new-password" || found.Token != "new-token" || found.IP != "new-ip" {
		t.Fatalf("credentials reverted: %+v", found)
	}
	if _, err := Authenticate(user.ID, "old-token"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("old token accepted: %v", err)
	}
}
func TestDeleteUserTreeIsAtomicAndRechecksActivity(t *testing.T) {
	resetData(t)
	user, network := createNormal(t, "delete-tree")
	db := storage.Get()
	route := Route{NetID: network.ID}
	if err := route.Create(); err != nil {
		t.Fatal(err)
	}
	device := Device{NetID: network.ID, VMac: "device", SessionID: "tree-session", IP: "192.168.202.1"}
	if err := device.Activate(); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE TRIGGER reject_user_delete BEFORE UPDATE OF deleted_at ON users BEGIN SELECT RAISE(ABORT, 'injected user delete failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Exec("DROP TRIGGER IF EXISTS reject_user_delete") })
	if _, err := DeleteUserTree(user.ID, nil); err == nil {
		t.Fatal("expected delete failure")
	}
	if _, err := GetNetByNetID(network.ID); err != nil {
		t.Fatalf("network deletion did not roll back: %v", err)
	}
	if _, err := GetRouteByRouteID(route.ID); err != nil {
		t.Fatalf("route deletion did not roll back: %v", err)
	}
	if _, err := GetDeviceByDevID(device.ID); err != nil {
		t.Fatalf("device deletion did not roll back: %v", err)
	}
	if err := db.Exec("DROP TRIGGER reject_user_delete").Error; err != nil {
		t.Fatal(err)
	}
	cutoff := time.Now().Add(-time.Hour)
	if ids, err := DeleteUserTree(user.ID, &cutoff); err != nil || len(ids) != 0 {
		t.Fatalf("active user cleanup: %v %v", ids, err)
	}
	if ids, err := DeleteUserTree(user.ID, nil); err != nil || len(ids) != 1 {
		t.Fatalf("delete: %v %v", ids, err)
	}
	if _, err := GetNetByNetID(network.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted net visible: %v", err)
	}
	if _, err := GetRouteByRouteID(route.ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted route visible: %v", err)
	}
}
func TestDeviceSessionsAndTrafficCannotGoBackwards(t *testing.T) {
	resetData(t)
	_, network := createNormal(t, "sessions")
	device := Device{NetID: network.ID, VMac: "test-device", SessionID: "old-session", IP: "192.168.202.1"}
	if err := device.Activate(); err != nil {
		t.Fatal(err)
	}
	device.RX = 200
	device.TX = 100
	if err := device.SaveTraffic(); err != nil {
		t.Fatal(err)
	}
	device.RX = 20
	device.TX = 10
	if err := device.SaveTraffic(); err != nil {
		t.Fatal(err)
	}
	old := device
	next := Device{NetID: network.ID, VMac: device.VMac, SessionID: "new-session", IP: device.IP}
	if err := next.Activate(); err != nil {
		t.Fatal(err)
	}
	if next.ID != old.ID || next.RX != 200 || next.TX != 100 {
		t.Fatalf("traffic not preserved: %+v", next)
	}
	old.Online = false
	if err := old.SaveRxTxOnline(); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("stale offline write: %v", err)
	}
	fresh, err := GetDeviceByDevID(next.ID)
	if err != nil || !fresh.Online || fresh.SessionID != "new-session" {
		t.Fatalf("new session changed: %+v %v", fresh, err)
	}
}
func TestConfigReadFailureIsNotDefault(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := GetConfig("openreg", "true", ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("query failure became default: %v", err)
	}
}
func TestUniqueKeysPermitSoftDeletedReuse(t *testing.T) {
	resetData(t)
	user, network := createNormal(t, "reuse")
	duplicate := Net{UserID: user.ID, Name: network.Name}
	if err := duplicate.Create(); !errors.Is(err, ErrNetworkExists) {
		t.Fatalf("duplicate network: %v", err)
	}
	if _, err := DeleteUserTree(user.ID, nil); err != nil {
		t.Fatal(err)
	}
	createNormal(t, "reuse")
	if err := SetConfig("reuse-config", "old"); err != nil {
		t.Fatal(err)
	}
	if err := DelConfig("reuse-config"); err != nil {
		t.Fatal(err)
	}
	if err := SetConfig("reuse-config", "new"); err != nil {
		t.Fatal(err)
	}
	if value, err := GetConfig("reuse-config", ""); err != nil || value != "new" {
		t.Fatalf("config reuse: %s %v", value, err)
	}
}

// The legacy device layout has exactly the original persisted fields.
type legacyDevice struct {
	gorm.Model
	NetID    uint
	VMac     string
	IP       string
	Online   bool
	RX       uint64
	TX       uint64
	OS       string
	Version  string
	Hostname string
	Country  string
	Region   string
}

func (legacyDevice) TableName() string { return "devices" }
func legacyDB(t *testing.T) (*gorm.DB, string) {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "legacy.db")
	db, err := storage.Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	pool, _ := db.DB()
	t.Cleanup(func() { pool.Close() })
	if err := db.AutoMigrate(&User{}, &Config{}, &Net{}, &legacyDevice{}, &Route{}); err != nil {
		t.Fatal(err)
	}
	return db, filename
}
func TestExistingDatabaseInitializationDoesNotChangeFileOrSchema(t *testing.T) {
	db, filename := legacyDB(t)
	rows := []User{{Name: "existing-duplicate"}, {Name: "existing-duplicate"}}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if err := InitializeSchema(db); err != nil {
		t.Fatal(err)
	}
	if err := InitializeSchema(db); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("opening an existing schema changed database bytes")
	}
	if db.Migrator().HasColumn(&Device{}, "session_id") {
		t.Fatal("session column was added")
	}
	if db.Migrator().HasTable("schema_migrations") {
		t.Fatal("migration table was added")
	}
	var indexes int64
	if err := db.Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND name LIKE 'uq_%'").Scan(&indexes).Error; err != nil || indexes != 0 {
		t.Fatalf("unique indexes added: %d %v", indexes, err)
	}
	var count int64
	if err := db.Model(&User{}).Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("existing duplicates modified: %d %v", count, err)
	}
}
func TestIncompleteExistingSchemaIsReportedWithoutChanges(t *testing.T) {
	db, filename := legacyDB(t)
	if err := db.Exec("ALTER TABLE devices DROP COLUMN hostname").Error; err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	err = InitializeSchema(db)
	if err == nil || !strings.Contains(err.Error(), "devices.hostname") {
		t.Fatalf("missing column not reported: %v", err)
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("incomplete schema was modified")
	}
}
func TestExistingAmbiguousRecordsAreReportedWithoutCleanup(t *testing.T) {
	resetData(t)
	db := storage.Get()
	if err := db.Create(&[]Config{{Key: "duplicate-key", Value: "one"}, {Key: "duplicate-key", Value: "two"}}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := GetConfig("duplicate-key", "default"); !errors.Is(err, ErrAmbiguousData) {
		t.Fatalf("ambiguous key: %v", err)
	}
	if err := SetConfig("duplicate-key", "replacement"); !errors.Is(err, ErrAmbiguousData) {
		t.Fatalf("ambiguous write: %v", err)
	}
	if err := db.Create(&[]User{{Name: "duplicate-user"}, {Name: "duplicate-user"}}).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := FindUser("duplicate-user"); !errors.Is(err, ErrAmbiguousData) {
		t.Fatalf("ambiguous username: %v", err)
	}
	var count int64
	if err := db.Model(&Config{}).Where("key = ?", "duplicate-key").Count(&count).Error; err != nil || count != 2 {
		t.Fatalf("duplicates modified: %d %v", count, err)
	}
}

func TestConcurrentConfigurationAndNetworkNamesWithoutUniqueIndexes(t *testing.T) {
	resetData(t)
	user, _ := createNormal(t, "operation-uniqueness")
	var wg sync.WaitGroup
	errs := make(chan error, 24)
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) { defer wg.Done(); errs <- SetConfig("concurrent-key", fmt.Sprint(i)) }(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int64
	if err := storage.Get().Model(&Config{}).Where("key = ?", "concurrent-key").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("configuration duplicated: %d %v", count, err)
	}
	errs = make(chan error, 24)
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); n := Net{UserID: user.ID, Name: "same-network"}; errs <- n.Create() }()
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		} else if !errors.Is(err, ErrNetworkExists) {
			t.Fatal(err)
		}
	}
	if success != 1 {
		t.Fatalf("network created %d times", success)
	}
	var indexes int64
	if err := storage.Get().Raw("SELECT count(*) FROM sqlite_master WHERE type='index' AND sql LIKE 'CREATE UNIQUE INDEX%'").Scan(&indexes).Error; err != nil || indexes != 0 {
		t.Fatalf("test relied on unique indexes: %d %v", indexes, err)
	}
}
func TestNormalWritesDoNotChangeDatabaseSchema(t *testing.T) {
	resetData(t)
	db := storage.Get()
	var before int64
	if err := db.Raw("PRAGMA schema_version").Scan(&before).Error; err != nil {
		t.Fatal(err)
	}
	user, network := createNormal(t, "unchanged-schema")
	device := Device{NetID: network.ID, VMac: "schema-device", IP: "192.168.202.1", SessionID: "memory-only"}
	if err := device.Activate(); err != nil {
		t.Fatal(err)
	}
	device.RX = 100
	if err := device.SaveTraffic(); err != nil {
		t.Fatal(err)
	}
	if err := SetConfig("schema-key", "value"); err != nil {
		t.Fatal(err)
	}
	if _, err := DeleteUserTree(user.ID, nil); err != nil {
		t.Fatal(err)
	}
	var after int64
	if err := db.Raw("PRAGMA schema_version").Scan(&after).Error; err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("business operations changed schema version: %d -> %d", before, after)
	}
}
