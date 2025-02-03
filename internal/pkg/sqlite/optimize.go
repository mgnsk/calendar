package sqlite

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// RunOptimizer runs an sqlite database periodic optimizer task.
func RunOptimizer(ctx context.Context, db *sql.DB) error {
	// Run thorough optimize on startup.
	if _, err := db.Exec(`PRAGMA optimize=0x10002`); err != nil {
		return err
	}

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			slog.Info("executing sqlite PRAGMA optimize")
			if _, err := db.Exec(`PRAGMA optimize`); err != nil {
				return err
			}
		}
	}
}
