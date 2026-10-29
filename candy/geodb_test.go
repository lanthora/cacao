package candy

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ipipdotnet/ipdb-go"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
	"github.com/oschwald/geoip2-golang"
)

// A small IPv4/IPv6 IPDB whose two root edges point to the same city record.
func testIPDB(t *testing.T) []byte {
	t.Helper()
	record := []byte("CN\tTest Region")
	data := make([]byte, 11+len(record))
	binary.BigEndian.PutUint32(data[:4], 2)
	binary.BigEndian.PutUint32(data[4:8], 2)
	binary.BigEndian.PutUint16(data[9:11], uint16(len(record)))
	copy(data[11:], record)
	meta, err := json.Marshal(ipdb.MetaData{IPVersion: 3, Languages: map[string]int{"CN": 0}, NodeCount: 1, TotalSize: len(data), Fields: []string{"country_code", "region_name"}})
	if err != nil {
		t.Fatal(err)
	}
	body := make([]byte, 4)
	binary.BigEndian.PutUint32(body, uint32(len(meta)))
	return append(append(body, meta...), data...)
}

func testBody(t *testing.T, kind string) []byte {
	t.Helper()
	if kind == "ipdb" {
		return testIPDB(t)
	}
	body, err := os.ReadFile("testdata/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

type memoryGeoConfig struct {
	mu     sync.Mutex
	values map[string]model.GeoDBConfig
	fail   bool
}

func testUpdater(t *testing.T, url string) (*GeoDBUpdater, *memoryGeoConfig) {
	t.Helper()
	u := NewGeoDBUpdater(t.TempDir())
	store := &memoryGeoConfig{values: map[string]model.GeoDBConfig{}}
	for _, kind := range []string{"ipdb", "mmdb"} {
		store.values[kind] = model.GeoDBConfig{Type: kind, DownloadURL: url, UpdateIntervalHours: 24}
	}
	u.getConfig = func(_ context.Context, kind string) (model.GeoDBConfig, error) {
		store.mu.Lock()
		defer store.mu.Unlock()
		return store.values[kind], nil
	}
	u.saveSuccess = func(_ context.Context, kind string, at time.Time) error {
		store.mu.Lock()
		defer store.mu.Unlock()
		if store.fail {
			return errors.New("simulated persistence failure")
		}
		c := store.values[kind]
		c.LastSuccessAt = at.Format(time.RFC3339Nano)
		store.values[kind] = c
		return nil
	}
	return u, store
}

func assertGeoQuery(t *testing.T, u *GeoDBUpdater, kind string) {
	t.Helper()
	filename, release, err := u.acquire(kind)
	defer release()
	if err != nil {
		t.Fatal(err)
	}
	if kind == "ipdb" {
		db, err := ipdb.NewCity(filename)
		if err != nil {
			t.Fatal(err)
		}
		info, err := db.FindInfo("1.1.1.1", "CN")
		if err != nil || info.CountryCode != "CN" {
			t.Fatalf("query: %+v %v", info, err)
		}
	} else {
		db, err := geoip2.Open(filename)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		info, err := db.City(net.ParseIP("81.2.69.160"))
		if err != nil || info.Country.IsoCode != "GB" {
			t.Fatalf("query: %+v %v", info, err)
		}
	}
}

func TestGeoDBPublishAndFailure(t *testing.T) {
	for _, kind := range []string{"ipdb", "mmdb"} {
		t.Run(kind, func(t *testing.T) {
			body := testBody(t, kind)
			var handler func(http.ResponseWriter, *http.Request)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler(w, r) }))
			defer server.Close()
			u, store := testUpdater(t, server.URL)
			handler = func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }
			legacy := filepath.Join(u.root, "legacy."+kind)
			if err := os.WriteFile(legacy, body, 0600); err != nil {
				t.Fatal(err)
			}
			if err := u.Update(context.Background(), kind); err != nil {
				t.Fatal(err)
			}
			first := u.states[kind].current
			assertGeoQuery(t, u, kind)
			if err := u.Update(context.Background(), kind); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(first); !os.IsNotExist(err) {
				t.Fatal("old managed version was not removed")
			}
			if _, err := os.Stat(legacy); err != nil {
				t.Fatal("legacy file removed", err)
			}
			current := u.states[kind].current
			config, _ := u.getConfig(context.Background(), kind)
			tests := map[string]func(http.ResponseWriter, *http.Request){
				"404":       func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) },
				"html":      func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "<html>error</html>") },
				"truncated": func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body[:len(body)/2]) },
				"oversized": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", fmt.Sprint(maxGeoDBSize+1))
				},
				"interrupted": func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", fmt.Sprint(len(body)+1))
					_, _ = w.Write(body[:4])
				},
			}
			for name, h := range tests {
				t.Run(name, func(t *testing.T) {
					handler = h
					if err := u.Update(context.Background(), kind); err == nil {
						t.Fatal("expected failure")
					}
					if u.states[kind].current != current {
						t.Fatal("current version changed")
					}
					c, _ := u.getConfig(context.Background(), kind)
					if c.LastSuccessAt != config.LastSuccessAt {
						t.Fatal("success timestamp changed")
					}
					entries, _ := os.ReadDir(u.directory(kind))
					if len(entries) != 1 {
						t.Fatalf("temporary files remain: %v", entries)
					}
					assertGeoQuery(t, u, kind)
				})
			}
			other := "ipdb"
			if kind == other {
				other = "mmdb"
			}
			store.mu.Lock()
			untouched := store.values[other].LastSuccessAt
			store.mu.Unlock()
			if untouched != "" {
				t.Fatal("other database timestamp changed")
			}
		})
	}
}

func TestGeoDBRecovery(t *testing.T) {
	for _, kind := range []string{"ipdb", "mmdb"} {
		t.Run(kind, func(t *testing.T) {
			body := testBody(t, kind)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
			defer server.Close()
			u, store := testUpdater(t, server.URL)
			store.fail = true
			if err := u.Update(context.Background(), kind); err == nil || !strings.Contains(err.Error(), "published") {
				t.Fatalf("expected partial success, got %v", err)
			}
			assertGeoQuery(t, u, kind)
			current := u.states[kind].current
			for name, content := range map[string][]byte{
				"20990101T000000.000000000Z." + kind: []byte("corrupt"),
				"20000101T000000.000000000Z." + kind: body,
				".download-interrupted.tmp":          []byte("partial"),
				"unmanaged.txt":                      []byte("keep"),
			} {
				if err := os.WriteFile(filepath.Join(u.directory(kind), name), content, 0600); err != nil {
					t.Fatal(err)
				}
			}
			store.fail = false
			restarted := NewGeoDBUpdater(u.root)
			restarted.getConfig = u.getConfig
			restarted.saveSuccess = u.saveSuccess
			if err := restarted.recoverType(kind); err != nil {
				t.Fatal(err)
			}
			if restarted.states[kind].current != current {
				t.Fatal("recovery selected wrong file")
			}
			c, _ := restarted.getConfig(context.Background(), kind)
			if c.LastSuccessAt == "" {
				t.Fatal("timestamp was not recovered")
			}
			entries, _ := os.ReadDir(u.directory(kind))
			if len(entries) != 2 {
				t.Fatalf("unexpected recovery files: %v", entries)
			}
			assertGeoQuery(t, restarted, kind)
		})
	}
}

func TestGeoDBConcurrentUpdatesAndReaders(t *testing.T) {
	started := make(chan string, 2)
	resume := make(chan struct{})
	bodies := map[string][]byte{"ipdb": testBody(t, "ipdb"), "mmdb": testBody(t, "mmdb")}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kind := strings.TrimPrefix(r.URL.Path, "/")
		started <- kind
		<-resume
		_, _ = w.Write(bodies[kind])
	}))
	defer server.Close()
	u, store := testUpdater(t, server.URL)
	for kind, c := range store.values {
		c.DownloadURL = server.URL + "/" + kind
		store.values[kind] = c
	}
	// Hold an existing query open while both downloads complete.
	_, release, _ := u.acquire("ipdb")
	results := make(chan error, 2)
	go func() { results <- u.Update(context.Background(), "ipdb") }()
	<-started
	if err := u.Update(context.Background(), "ipdb"); !errors.Is(err, ErrGeoDBBusy) {
		t.Fatalf("duplicate update: %v", err)
	}
	go func() { results <- u.Update(context.Background(), "mmdb") }()
	<-started
	close(resume)
	// MMDB can publish while IPDB waits for its reader to close.
	if err := <-results; err != nil {
		release()
		t.Fatal(err)
	}
	select {
	case err := <-results:
		release()
		t.Fatalf("published while reader held lock: %v", err)
	default:
	}
	release()
	if err := <-results; err != nil {
		t.Fatal(err)
	}
	assertGeoQuery(t, u, "ipdb")
	assertGeoQuery(t, u, "mmdb")
}

func TestGeoDBDue(t *testing.T) {
	now := time.Now().UTC()
	c := model.GeoDBConfig{Type: "ipdb", DownloadURL: "https://example.com/city.ipdb", UpdateIntervalHours: 24}
	if !geoDBDue(c, now) {
		t.Fatal("initial update not due")
	}
	c.LastSuccessAt = now.Add(-24 * time.Hour).Format(time.RFC3339Nano)
	if !geoDBDue(c, now) {
		t.Fatal("boundary not due")
	}
	c.LastSuccessAt = now.Add(-23 * time.Hour).Format(time.RFC3339Nano)
	if geoDBDue(c, now) {
		t.Fatal("updated too early")
	}
	c.UpdateIntervalHours = 0
	if geoDBDue(c, now) {
		t.Fatal("disabled interval")
	}
	c.UpdateIntervalHours = 24
	c.DownloadURL = ""
	if geoDBDue(c, now) {
		t.Fatal("empty URL")
	}
}

func TestGeoDBDownloadCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	u, _ := testUpdater(t, server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := u.Update(ctx, "ipdb"); err == nil {
		t.Fatal("expected timeout")
	}
	entries, _ := os.ReadDir(u.directory("ipdb"))
	if len(entries) != 0 {
		t.Fatal("temporary file remains")
	}
}

func TestGeoDBLegacyFallback(t *testing.T) {
	u, _ := testUpdater(t, "")
	for _, kind := range []string{"ipdb", "mmdb"} {
		if err := os.WriteFile(filepath.Join(u.root, "legacy."+kind), testBody(t, kind), 0600); err != nil {
			t.Fatal(err)
		}
		assertGeoQuery(t, u, kind)
	}
}

func TestGeoDBScheduledAndManualUpdates(t *testing.T) {
	for _, kind := range []string{"ipdb", "mmdb"} {
		t.Run(kind, func(t *testing.T) {
			body := testBody(t, kind)
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests++; _, _ = w.Write(body) }))
			defer server.Close()
			u, store := testUpdater(t, server.URL)
			if err := u.update(context.Background(), kind, true); err != nil {
				t.Fatal(err)
			}
			if err := u.update(context.Background(), kind, true); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatalf("scheduled update ignored interval: %d", requests)
			}
			store.mu.Lock()
			c := store.values[kind]
			c.UpdateIntervalHours = 0
			store.values[kind] = c
			store.mu.Unlock()
			if err := u.update(context.Background(), kind, true); err != nil {
				t.Fatal(err)
			}
			if requests != 1 {
				t.Fatal("disabled update downloaded")
			}
			if err := u.Update(context.Background(), kind); err != nil {
				t.Fatal(err)
			}
			if requests != 2 {
				t.Fatal("manual update obeyed automatic interval")
			}
		})
	}
}

func TestGeoDBRejectMalformedIPDB(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{"truncated header", func(b []byte) []byte { return b[:3] }},
		{"metadata size", func(b []byte) []byte { binary.BigEndian.PutUint32(b, ^uint32(0)); return b }},
		{"record pointer", func(b []byte) []byte {
			start := 4 + int(binary.BigEndian.Uint32(b))
			binary.BigEndian.PutUint32(b[start:], ^uint32(0))
			return b
		}},
		{"record length", func(b []byte) []byte {
			start := 4 + int(binary.BigEndian.Uint32(b))
			binary.BigEndian.PutUint16(b[start+9:], 65535)
			return b
		}},
		{"empty node", func(b []byte) []byte {
			start := 4 + int(binary.BigEndian.Uint32(b))
			binary.BigEndian.PutUint32(b[start:], 1)
			return b
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "malformed.ipdb")
			if err := os.WriteFile(filename, test.mutate(testIPDB(t)), 0600); err != nil {
				t.Fatal(err)
			}
			if err := validateGeoDB("ipdb", filename); err == nil {
				t.Fatal("accepted malformed database")
			}
		})
	}
}

func TestGeoDBSQLitePersistenceRecovery(t *testing.T) {
	body := testBody(t, "ipdb")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(body) }))
	defer server.Close()
	if err := model.SaveGeoDBConfig(model.GeoDBConfig{Type: "ipdb", DownloadURL: server.URL, UpdateIntervalHours: 24}); err != nil {
		t.Fatal(err)
	}
	db := storage.Get()
	if err := db.Exec("CREATE TRIGGER reject_geodb_time BEFORE INSERT ON configs WHEN NEW.key = 'ipdbLastSuccessAt' BEGIN SELECT RAISE(ABORT, 'injected timestamp failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Exec("DROP TRIGGER IF EXISTS reject_geodb_time")
		for _, key := range []string{"ipdbDownloadURL", "ipdbUpdateIntervalHours", "ipdbLastSuccessAt"} {
			if err := model.DelConfig(key); err != nil {
				t.Error(err)
			}
		}
	})
	updater := NewGeoDBUpdater(t.TempDir())
	if err := updater.Update(context.Background(), "ipdb"); err == nil || !strings.Contains(err.Error(), "published") {
		t.Fatalf("expected partial success: %v", err)
	}
	assertGeoQuery(t, updater, "ipdb")
	config, err := model.GetGeoDBConfig("ipdb")
	if err != nil || config.LastSuccessAt != "" {
		t.Fatalf("failed timestamp persisted: %+v %v", config, err)
	}
	if err := db.Exec("DROP TRIGGER reject_geodb_time").Error; err != nil {
		t.Fatal(err)
	}
	restarted := NewGeoDBUpdater(updater.root)
	if err := restarted.recoverType("ipdb"); err != nil {
		t.Fatal(err)
	}
	assertGeoQuery(t, restarted, "ipdb")
	config, err = model.GetGeoDBConfig("ipdb")
	if err != nil || config.LastSuccessAt == "" {
		t.Fatalf("timestamp not recovered: %+v %v", config, err)
	}
}
