package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar/pkg/snowflake"
)

// Invite is the invite domain model.
type Invite struct {
	Token      uuid.UUID
	ValidUntil time.Time
	CreatedBy  snowflake.ID
}
