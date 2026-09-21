package migrate

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"workspace/migrations"
)

func Up(db *gorm.DB) error {
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version text PRIMARY KEY,
			applied_at timestamptz NOT NULL DEFAULT now()
		)
	`).Error; err != nil {
		return fmt.Errorf("schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations.SQL, ".")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var n int64
		if err := db.Raw(`SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, name).Scan(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		body, err := fs.ReadFile(migrations.SQL, name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(body)).Error; err != nil {
				return err
			}
			return tx.Exec(
				`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`,
				name, time.Now().UTC(),
			).Error
		})
		if err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
	}
	return nil
}
