package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	sqliteDriver "github.com/glebarez/go-sqlite"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var db *gorm.DB

// Initialize is explicit: importing model/storage never opens a production file.
func Initialize(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var err error
	db, err = Open(filepath.Join(dir, "sqlite.db"))
	return err
}

// Open configures every connection, including connections reopened by database/sql.
// A single connection serializes this process; IMMEDIATE transactions also protect
// read-then-write operations from competing processes. WAL is intentionally opt-in.
func Open(filename string) (*gorm.DB, error) {
	absolute, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	normalized := filepath.ToSlash(absolute)
	if filepath.VolumeName(absolute) != "" && !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	uri := (&url.URL{Scheme: "file", Path: normalized}).String()
	conn, err := gorm.Open(sqlite.Open(uri+"?_pragma=busy_timeout(100)&_txlock=immediate"), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		return nil, err
	}
	pool, err := conn.DB()
	if err != nil {
		return nil, err
	}
	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)
	return conn, nil
}

func Get() *gorm.DB { return db }

// Transaction callbacks may be replayed: use only tx for database work, reset
// attempt-local models, and perform external/visible side effects after success.
// No callback may acquire application locks or re-enter the global connection.
func Transaction(ctx context.Context, fn func(*gorm.DB) error) error {
	return TransactionOn(ctx, db, fn)
}
func TransactionOn(parent context.Context, conn *gorm.DB, fn func(*gorm.DB) error) error {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	var err error
	for attempt := 0; attempt < 8; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err = transactionAttempt(ctx, conn, fn)
		if !IsBusy(err) {
			return err
		}
		if attempt == 7 {
			break
		}
		delay := time.Duration(10*(1<<min(attempt, 5))) * time.Millisecond
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf("database busy after bounded transaction retries: %w", err)
}

// Keep ownership of the physical connection through rollback. This driver's
// database/sql Tx.Commit leaves SQLite's transaction open on SQLITE_BUSY, while
// database/sql already marks the Tx done; explicit connection ownership avoids
// returning a still-locked connection to the pool after a failed commit.
func transactionAttempt(ctx context.Context, db *gorm.DB, fn func(*gorm.DB) error) error {
	return db.WithContext(ctx).Connection(func(connection *gorm.DB) (err error) {
		tx := connection.Session(&gorm.Session{SkipDefaultTransaction: true})
		if err = tx.Exec("BEGIN IMMEDIATE").Error; err != nil {
			return err
		}
		committed := false
		defer func() {
			if !committed {
				cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				if rollbackErr := tx.WithContext(cleanup).Exec("ROLLBACK").Error; rollbackErr != nil {
					if physical, ok := tx.Statement.ConnPool.(*sql.Conn); ok {
						_ = physical.Raw(func(interface{}) error { return driver.ErrBadConn })
					}
					err = fmt.Errorf("transaction failed (%v); rollback failed: %w", err, rollbackErr)
				}
			}
		}()
		if err = fn(tx); err != nil {
			return err
		}
		err = tx.Exec("COMMIT").Error
		committed = err == nil
		return err
	})
}

func IsBusy(err error) bool {
	var e *sqliteDriver.Error
	return errors.As(err, &e) && e.Code()&255 == 5
}
func IsUnique(err error) bool {
	var e *sqliteDriver.Error
	return errors.As(err, &e) && (e.Code() == 2067 || e.Code() == 1555)
}
