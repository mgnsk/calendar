package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// NormalizeError wraps the error with additional information if present.
func NormalizeError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return wreck.NotFound.New("Not found", err)
	}

	if se := new(sqlite.Error); errors.As(err, &se) {
		switch code := se.Code(); code {
		case
			sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY,
			sqlite3.SQLITE_CONSTRAINT_UNIQUE:
			return wreck.AlreadyExists.New("Already exists", err)

		case sqlite3.SQLITE_LOCKED:
			return wreck.Timeout.New("Timeout", err)

		default:
			return fmt.Errorf("code %s (%d): %w", sqlite.ErrorCodeString[code], code, err)
		}
	}

	return err
}

// WithErrorChecking handles executed query errors.
func WithErrorChecking(res sql.Result, err error) error {
	if err != nil {
		return NormalizeError(err)
	}

	if c, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("error checking affected row count: %w", err)
	} else if c == 0 {
		return wreck.PreconditionFailed.New("Now rows affected", err)
	}

	return nil
}
