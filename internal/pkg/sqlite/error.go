package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/mgnsk/calendar/internal/pkg/wreck"
)

// WithErrorChecking handles executed query errors.
func WithErrorChecking(res sql.Result, err error) error {
	if err != nil {
		return err
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
