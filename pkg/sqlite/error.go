package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mgnsk/calendar"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// NormalizeError wraps the error with additional information if present.
func NormalizeError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return calendar.NotFound.New("Not found", err)
	}

	var se *sqlite.Error
	if errors.As(err, &se) {
		switch se.Code() {
		case
			sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY,
			sqlite3.SQLITE_CONSTRAINT_UNIQUE:
			return calendar.AlreadyExists.New("Already exists", err)

		case sqlite3.SQLITE_LOCKED:
			return calendar.Timeout.New("Timeout", err)

		case sqlite3.SQLITE_ERROR:
			if strings.Contains(se.Error(), "fts5: syntax error") {
				return calendar.InvalidValue.New("Invalid query", err)
			}
		}

		return fmt.Errorf("code %s (%d): %w", sqlite.ErrorCodeString[se.Code()], se.Code(), err)
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
		return calendar.PreconditionFailed.New("Now rows affected", err)
	}

	return nil
}
