package domain

import (
	"time"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
)

// Event is the event domain model.
type Event struct {
	ID          snowflake.ID
	StartAt     timestamp.Timestamp
	EndAt       timestamp.Timestamp
	Title       string
	Description string
	URL         string
	Tags        []string
}

// GetCreatedAt returns the event created at time.
func (e *Event) GetCreatedAt() time.Time {
	return snowflake.ParseTime(e.ID.Int64())
}
