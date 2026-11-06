package calendar

import (
	"database/sql"
	"embed"

	"github.com/mgnsk/webkit/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrateUp runs the up migrations for database.
func MigrateUp(db *sql.DB) error {
	return sqlite.MigrateUp(db, migrationsFS, "migrations")
}

// MigrateDown runs the down migrations for database.
func MigrateDown(db *sql.DB) error {
	return sqlite.MigrateDown(db, migrationsFS, "migrations")
}
