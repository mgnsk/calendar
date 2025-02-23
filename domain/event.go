package domain

import (
	"slices"
	"strings"
	"time"

	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/textfilter"
)

// Event is the event domain model.
type Event struct {
	ID          snowflake.ID
	StartAt     time.Time
	Title       string
	Description string
	URL         string
	Location    string
	Latitude    float64
	Longitude   float64
	IsDraft     bool
	UserID      snowflake.ID
}

// GetCreatedAt returns the event created at time.
func (e *Event) GetCreatedAt() time.Time {
	return snowflake.ParseTime(e.ID.Int64())
}

// GetDateString returns a formatted string with event start and end times.
// TODO: handle multi-day dates
func (e *Event) GetDateString() string {
	var buf strings.Builder
	buf.WriteString(e.StartAt.Format("January _2, 2006 "))

	if e.StartAt.Minute() == 0 {
		buf.WriteString(e.StartAt.Format("3PM"))
	} else {
		buf.WriteString(e.StartAt.Format("3:04PM"))
	}

	return buf.String()
}

// GetTags returns unique words in title and description.
// A word is defined as having at least 3 characters.
func (e *Event) GetTags() []string {
	var words []string
	for _, source := range []string{e.Title, e.Description, e.Location} {
		words = append(words, textfilter.GetTags(source)...)
	}

	slices.Sort(words)
	words = slices.Compact(words)

	return words
}
