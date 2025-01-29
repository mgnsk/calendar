package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"modernc.org/sqlite"
)

// DriverError is a database driver error.
type DriverError interface {
	Code() int
	error
}

var _ DriverError = &sqlite.Error{}

// NormalizeError wraps the error with additional information if present.
func NormalizeError(err error) error {
	if err == nil {
		return nil
	}

	var se DriverError
	if errors.As(err, &se) {
		return fmt.Errorf("sqlite error %v (%s): %w", se.Code(), sqlite.ErrorCodeString[se.Code()], err)
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
		return &wreck.PreconditionFailed{
			Cause: fmt.Errorf("no rows were affected"),
		}
	}

	return nil
}
