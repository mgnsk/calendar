package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// RunOptimizer runs an sqlite database periodic optimizer task.
func RunOptimizer(ctx context.Context, db *sql.DB) error {
	slog.Info("optimizing sqlite database")
	start := time.Now()

	// Run thorough optimize on startup.
	if _, err := db.Exec(`PRAGMA optimize=0x10002`); err != nil {
		return err
	}

	if _, err := db.Exec(`INSERT INTO events_fts(events_fts) VALUES('optimize')`); err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("database optimization finished in %v", time.Since(start)))

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil

		case now := <-ticker.C:
			slog.Info("optimizing sqlite database")

			if _, err := db.Exec(`PRAGMA optimize`); err != nil {
				return err
			}
			if _, err := db.Exec(`INSERT INTO events_fts(events_fts) VALUES('optimize')`); err != nil {
				return err
			}

			slog.Info(fmt.Sprintf("database optimization finished in %v", time.Since(now)))
		}
	}
}
