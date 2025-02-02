package domain

import "github.com/mgnsk/calendar/internal/pkg/snowflake"

// Tag is the tag domain model.
type Tag struct {
	ID         snowflake.ID
	Name       string
	EventCount uint64
}
