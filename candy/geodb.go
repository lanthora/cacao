package candy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/util"
)

const geoDBVersionLayout = "20060102T150405.000000000Z"
const maxGeoDBSize int64 = 256 << 20

var ErrGeoDBBusy = errors.New("database update already in progress")

type geoDBState struct {
	update  sync.Mutex
	query   sync.RWMutex
	current string
	pending *time.Time
}

type GeoDBUpdater struct {
	root        string
	client      *http.Client
	states      map[string]*geoDBState
	validate    func(string, string) error
	getConfig   func(string) (model.GeoDBConfig, error)
	saveSuccess func(string, time.Time) error
}

func NewGeoDBUpdater(root string) *GeoDBUpdater {
	return &GeoDBUpdater{
		root: root,
		client: &http.Client{Timeout: 5 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many download redirects")
			}
			if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
				return errors.New("unsupported redirect scheme")
			}
			if via[len(via)-1].URL.Scheme == "https" && req.URL.Scheme != "https" {
				return errors.New("insecure download redirect")
			}
			return nil
		}},
		states:   map[string]*geoDBState{"ipdb": {}, "mmdb": {}},
		validate: validateGeoDB, getConfig: model.GetGeoDBConfig, saveSuccess: model.SaveGeoDBSuccess,
	}
}

var GeoDatabases = NewGeoDBUpdater(argp.Get("storage", "."))

func (u *GeoDBUpdater) directory(kind string) string { return filepath.Join(u.root, "geodb", kind) }

func versionTime(name, kind string) (time.Time, error) {
	return time.Parse(geoDBVersionLayout, strings.TrimSuffix(name, "."+kind))
}

// recoverType is called under the update lock, before serving or downloading.
func (u *GeoDBUpdater) recoverType(kind string) error {
	s := u.states[kind]
	s.query.Lock()
	defer s.query.Unlock()
	entries, err := os.ReadDir(u.directory(kind))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() > entries[j].Name() })
	previousCurrent := s.current
	s.current = ""
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(name, ".download-") && strings.HasSuffix(name, ".tmp") {
			if err := os.Remove(filepath.Join(u.directory(kind), name)); err != nil {
				logger.Warn("remove temporary geodb: %v", err)
			}
			continue
		}
		at, err := versionTime(name, kind)
		if err != nil || filepath.Ext(name) != "."+kind || s.current != "" {
			continue
		}
		filename := filepath.Join(u.directory(kind), name)
		if filename != previousCurrent {
			if err := u.validate(kind, filename); err != nil {
				logger.Warn("ignore invalid %s version: %v", kind, err)
				continue
			}
		}
		s.current = filename
		// A published filename is durable evidence of success even if the process
		// stopped between publishing it and saving the database timestamp.
		c, err := u.getConfig(kind)
		if err != nil {
			return err
		}
		previous, _ := time.Parse(time.RFC3339Nano, c.LastSuccessAt)
		if at.After(previous) {
			if err := u.saveSuccess(kind, at); err != nil {
				s.pending = &at
				return err
			}
			s.pending = nil
		}
	}
	if s.current != "" {
		u.cleanup(kind, s.current)
	}
	return nil
}

// Start recovers published versions synchronously; downloads run in the background.
func (u *GeoDBUpdater) Start(ctx context.Context) {
	for _, kind := range []string{"ipdb", "mmdb"} {
		s := u.states[kind]
		s.update.Lock()
		if err := u.recoverType(kind); err != nil {
			logger.Warn("recover %s: %v", kind, err)
		}
		s.update.Unlock()
		go func(kind string) {
			u.check(ctx, kind)
			ticker := time.NewTicker(time.Hour)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					u.check(ctx, kind)
				}
			}
		}(kind)
	}
}

func (u *GeoDBUpdater) check(ctx context.Context, kind string) {
	if err := u.update(ctx, kind, true); err != nil && !errors.Is(err, ErrGeoDBBusy) {
		logger.Warn("update %s: %v", kind, err)
	}
}

func (u *GeoDBUpdater) Update(ctx context.Context, kind string) error {
	return u.update(ctx, kind, false)
}

func (u *GeoDBUpdater) update(ctx context.Context, kind string, scheduled bool) error {
	s, ok := u.states[kind]
	if !ok {
		return errors.New("unsupported database type")
	}
	if !s.update.TryLock() {
		return ErrGeoDBBusy
	}
	defer s.update.Unlock()
	// Retry persistence without re-scanning or blocking readers during download.
	if s.pending != nil {
		if err := u.saveSuccess(kind, *s.pending); err != nil {
			return err
		}
		s.pending = nil
	}
	c, err := u.getConfig(kind)
	if err != nil {
		return err
	}
	if err := c.Validate(); err != nil {
		return err
	}
	if scheduled && !geoDBDue(c, time.Now()) {
		s.query.Lock()
		if s.current != "" {
			u.cleanup(kind, s.current)
		}
		s.query.Unlock()
		return nil
	}
	if c.DownloadURL == "" {
		return errors.New("download URL is empty")
	}
	if err := os.MkdirAll(u.directory(kind), 0755); err != nil {
		return err
	}
	file, err := os.CreateTemp(u.directory(kind), ".download-*.tmp")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	defer file.Close()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.DownloadURL, nil)
	if err != nil {
		return err
	}
	resp, err := u.client.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxGeoDBSize {
		return errors.New("database exceeds 256 MiB")
	}
	n, err := io.Copy(file, io.LimitReader(resp.Body, maxGeoDBSize+1))
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	if n == 0 || n > maxGeoDBSize {
		return errors.New("database is empty or exceeds 256 MiB")
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := u.validate(kind, temporary); err != nil {
		return fmt.Errorf("invalid %s database: %w", kind, err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	at := time.Now().UTC()
	filename := filepath.Join(u.directory(kind), at.Format(geoDBVersionLayout)+"."+kind)
	s.query.Lock()
	defer s.query.Unlock()
	if err := os.Rename(temporary, filename); err != nil {
		return err
	}
	s.current = filename
	// Directory sync is not available on every supported platform (e.g. Windows).
	if dir, err := os.Open(u.directory(kind)); err == nil {
		_ = dir.Sync()
		_ = dir.Close()
	}
	s.pending = &at
	if err := u.saveSuccess(kind, at); err != nil {
		return fmt.Errorf("database published, but saving update time failed: %w", err)
	}
	s.pending = nil
	u.cleanup(kind, filename)
	logger.Info("updated %s database", kind)
	return nil
}

func geoDBDue(c model.GeoDBConfig, now time.Time) bool {
	if c.DownloadURL == "" || c.UpdateIntervalHours <= 0 {
		return false
	}
	last, err := time.Parse(time.RFC3339Nano, c.LastSuccessAt)
	return err != nil || now.Sub(last) >= time.Duration(c.UpdateIntervalHours)*time.Hour
}

func (u *GeoDBUpdater) cleanup(kind, current string) {
	entries, err := os.ReadDir(u.directory(kind))
	if err != nil {
		logger.Warn("list old %s versions: %v", kind, err)
		return
	}
	for _, e := range entries {
		filename := filepath.Join(u.directory(kind), e.Name())
		if e.IsDir() || filename == current || filepath.Ext(e.Name()) != "."+kind {
			continue
		}
		if _, err := versionTime(e.Name(), kind); err != nil {
			continue
		}
		if err := os.Remove(filename); err != nil {
			logger.Warn("remove old %s version: %v", kind, err)
		}
	}
}

// The caller must defer release until its reader is closed.
func (u *GeoDBUpdater) acquire(kind string) (filename string, release func(), err error) {
	s := u.states[kind]
	s.query.RLock()
	release = s.query.RUnlock
	if s.current != "" {
		return s.current, release, nil
	}
	name, err := util.FindFileByExtFromDir(u.root, "."+kind)
	return filepath.Join(u.root, name), release, err
}
