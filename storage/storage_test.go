package storage

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"
)

func testDatabase(t *testing.T, filename string) *gorm.DB {
	t.Helper()
	db, err := Open(filename)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return db
}
func TestConcurrentReadThenWrite(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "concurrent.db")
	a := testDatabase(t, filename)
	b := testDatabase(t, filename)
	if err := a.Exec("CREATE TABLE counter (value INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	if err := a.Exec("INSERT INTO counter VALUES (0)").Error; err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			db := a
			if i%2 == 0 {
				db = b
			}
			errs <- TransactionOn(context.Background(), db, func(tx *gorm.DB) error {
				var value int
				if err := tx.Raw("SELECT value FROM counter").Scan(&value).Error; err != nil {
					return err
				}
				return tx.Exec("UPDATE counter SET value = ?", value+1).Error
			})
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var value int
	if err := a.Raw("SELECT value FROM counter").Scan(&value).Error; err != nil || value != 40 {
		t.Fatalf("counter %d: %v", value, err)
	}
}
func TestCommitBusyRollsBackAndReplays(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "commit.db")
	writer := testDatabase(t, filename)
	reader := testDatabase(t, filename)
	if err := writer.Exec("CREATE TABLE counter (value INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	if err := writer.Exec("INSERT INTO counter VALUES (0)").Error; err != nil {
		t.Fatal(err)
	}
	pool, _ := reader.DB()
	readTx, err := pool.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer readTx.Rollback()
	var value int
	if err := readTx.QueryRow("SELECT value FROM counter").Scan(&value); err != nil {
		t.Fatal(err)
	}
	var attempts atomic.Int32
	entered := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		result <- TransactionOn(context.Background(), writer, func(tx *gorm.DB) error {
			if attempts.Add(1) == 1 {
				close(entered)
			}
			return tx.Exec("UPDATE counter SET value = value + 1").Error
		})
	}()
	<-entered
	// The reader keeps a SHARED lock: BEGIN/UPDATE succeeds, COMMIT must retry.
	deadline := time.Now().Add(2 * time.Second)
	for attempts.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if attempts.Load() < 2 {
		t.Fatal("commit contention did not replay the transaction")
	}
	if err := readTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	if err := writer.Raw("SELECT value FROM counter").Scan(&value).Error; err != nil || value != 1 {
		t.Fatalf("rollback/replay produced %d, %v", value, err)
	}
	if err := TransactionOn(context.Background(), writer, func(tx *gorm.DB) error { return tx.Exec("UPDATE counter SET value = 2").Error }); err != nil {
		t.Fatalf("connection left inside transaction: %v", err)
	}
}
func TestTransactionCancellationAndNonBusyErrors(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "cancel.db")
	writer := testDatabase(t, filename)
	blocker := testDatabase(t, filename)
	if err := writer.Exec("CREATE TABLE counter (value INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	pool, _ := blocker.DB()
	lock, err := pool.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	start := time.Now()
	err = TransactionOn(ctx, writer, func(tx *gorm.DB) error { return tx.Exec("INSERT INTO counter VALUES (1)").Error })
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancellation: %v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("cancellation waited too long")
	}
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	expected := errors.New("business failure")
	calls := 0
	err = TransactionOn(context.Background(), writer, func(tx *gorm.DB) error {
		calls++
		if err := tx.Exec("INSERT INTO counter VALUES (2)").Error; err != nil {
			return err
		}
		return expected
	})
	if !errors.Is(err, expected) || calls != 1 {
		t.Fatalf("nonbusy retry: calls=%d err=%v", calls, err)
	}
	var count int64
	if err := writer.Raw("SELECT count(*) FROM counter").Scan(&count).Error; err != nil || count != 0 {
		t.Fatalf("failed writes committed: %d %v", count, err)
	}
}
func TestPanicRollsBack(t *testing.T) {
	db := testDatabase(t, filepath.Join(t.TempDir(), "panic.db"))
	if err := db.Exec("CREATE TABLE counter (value INTEGER)").Error; err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("panic swallowed")
			}
		}()
		_ = TransactionOn(context.Background(), db, func(tx *gorm.DB) error {
			if err := tx.Exec("INSERT INTO counter VALUES (1)").Error; err != nil {
				return err
			}
			panic("abort")
		})
	}()
	var count int64
	if err := db.Raw("SELECT count(*) FROM counter").Scan(&count).Error; err != nil || count != 0 {
		t.Fatalf("panic cleanup: %d %v", count, err)
	}
}
