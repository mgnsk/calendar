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

// IsValid returns whether the invite is valid.
func (i *Invite) IsValid() bool {
	return time.Until(i.ValidUntil) > 0
}
