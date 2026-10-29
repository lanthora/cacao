package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

// Initialize creates the original tables for a new installation. Existing files
// are inspected only: no AutoMigrate, new columns, indexes or migration records.
func Initialize() error { return InitializeSchema(storage.Get()) }
func InitializeSchema(db *gorm.DB) error {
	tables, err := db.Migrator().GetTables()
	if err != nil {
		return err
	}
	models := []interface{}{&Config{}, &User{}, &Net{}, &Route{}, &Device{}}
	existing := false
	for _, table := range tables {
		if !strings.HasPrefix(table, "sqlite_") {
			existing = true
			break
		}
	}
	if !existing {
		return storage.TransactionOn(context.Background(), db, func(tx *gorm.DB) error { return tx.AutoMigrate(models...) })
	}
	for _, value := range models {
		statement := &gorm.Statement{DB: db}
		if err := statement.Parse(value); err != nil {
			return err
		}
		columns, err := db.Migrator().ColumnTypes(value)
		if err != nil {
			return fmt.Errorf("inspect existing table %s: %w", statement.Schema.Table, err)
		}
		found := map[string]bool{}
		for _, column := range columns {
			found[column.Name()] = true
		}
		for _, name := range statement.Schema.DBNames {
			if !found[name] {
				return fmt.Errorf("existing database is missing %s.%s; schema left unchanged", statement.Schema.Table, name)
			}
		}
	}
	return nil
}
