package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"

	// Register the "sqlite" driver.
	_ "modernc.org/sqlite"
)

// Builder is an SQLite client builder.
type Builder struct {
	dsn   string
	debug bool
}

// NewDB creates an SQLite client.
func NewDB(dsn string) *Builder {
	return &Builder{
		dsn: dsn,
	}
}

// WithDebugLogging configures the client with query logging.
func (c *Builder) WithDebugLogging() *Builder {
	c.debug = true
	return c
}

// Connect to the database.
func (c *Builder) Connect() *bun.DB {
	sqldb, err := sql.Open("sqlite", c.dsn)
	if err != nil {
		panic(fmt.Errorf("error opening DB: %s", err))
	}

	sqldb.SetMaxIdleConns(1)
	sqldb.SetMaxOpenConns(1)

	db := bun.NewDB(sqldb, sqlitedialect.New())

	if c.debug {
		db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("error connecting to DB: %s", err))
	}

	return db
}
