package timestamp

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// New creates a new RFC3339Time.
func New(ts time.Time) Timestamp {
	return Timestamp{ts}
}

// Timestamp is an RFC3339 timestamp value for use in SQLite's string column.
// It's marshaled zero value is an empty string.
type Timestamp struct {
	value time.Time
}

// Time returns the time.Time value.
func (t Timestamp) Time() time.Time {
	return t.value
}

// String returns the value as string.
func (t Timestamp) String() string {
	if t.value.IsZero() {
		return ""
	}

	return t.value.Format(time.RFC3339)
}

// Scan implements sql.Scanner interface.
func (t *Timestamp) Scan(src interface{}) error {
	if src == nil {
		return nil
	}

	var val string

	switch src := src.(type) {
	case string:
		val = src

	default:
		panic(fmt.Sprintf("expected string, got %T", src))
	}

	if val == "" {
		return nil
	}

	ts, err := time.Parse(time.RFC3339, val)
	if err != nil {
		return err
	}

	t.value = ts

	return nil
}

// Value implements sql.Valuer interface.
func (t Timestamp) Value() (driver.Value, error) {
	if t.value.IsZero() {
		return "", nil
	}

	return t.value.Format(time.RFC3339), nil
}
