package internal

import (
	"database/sql"
	"embed"
	"errors"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrateUp runs the up migrations for database.
func MigrateUp(db *sql.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			slog.Info("migrations already up to date")
			return nil
		}

		return err
	}

	return nil
}

// MigrateDown runs the down migrations for database.
func MigrateDown(db *sql.DB) error {
	m, err := newMigrator(db)
	if err != nil {
		return err
	}

	return m.Down()
}

func newMigrator(db *sql.DB) (*migrate.Migrate, error) {
	sourceInstance, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}

	dbInstance, err := migratesqlite.WithInstance(db, &migratesqlite.Config{})
	if err != nil {
		return nil, err
	}

	return migrate.NewWithInstance("iofs", sourceInstance, "sqlite", dbInstance)
}
