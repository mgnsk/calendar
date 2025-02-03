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

// Timestamp is an nullable RFC3339 timestamp value for use in SQLite's string column.
// It's marshaled zero value is nil.
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
		return nil, nil
	}

	return t.value.Format(time.RFC3339), nil
}

// FormatDay returns day with the ordinal suffix for day.
func FormatDay(day int) string {
	return fmt.Sprintf("%d%s", day, getDaySuffix(day))
}

func getDaySuffix(n int) string {
	if n >= 11 && n <= 13 {
		return "th"
	}

	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}
