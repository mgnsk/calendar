package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/textfilter"
)

// Event is the event domain model.
type Event struct {
	ID          snowflake.ID
	StartAt     time.Time
	EndAt       time.Time
	Title       string
	Description string
	URL         string
}

// GetCreatedAt returns the event created at time.
func (e *Event) GetCreatedAt() time.Time {
	return snowflake.ParseTime(e.ID.Int64())
}

func (e *Event) GetDateString() string {
	var buf strings.Builder
	buf.WriteString(e.StartAt.Format("January _2, 2006 "))

	if e.StartAt.Minute() == 0 {
		buf.WriteString(e.StartAt.Format("3PM"))
	} else {
		buf.WriteString(e.StartAt.Format("3:04PM"))
	}

	if !e.EndAt.IsZero() {
		buf.WriteString("-")
		if e.EndAt.Minute() == 0 {
			buf.WriteString(e.EndAt.Format("3PM"))
		} else {
			buf.WriteString(e.EndAt.Format("3:04PM"))
		}
	}

	return buf.String()
}

// GetTags returns unique words in title and description.
// A word is defined as having at least 3 characters.
func (e *Event) GetTags() []string {
	var words []string
	for _, source := range []string{e.Title, e.Description} {
		words = append(words, textfilter.GetTags(source)...)
	}

	slices.Sort(words)
	words = slices.Compact(words)

	return words
}

// GetDescription returns the event description with tags.
// TODO: test this
func (e *Event) GetDescription() string {
	var buf strings.Builder

	buf.WriteString(e.Description)

	// TODO: get top tags
	// if len(e.Tags) > 0 {
	// 	buf.WriteString(fmt.Sprintf("\n\ntags: %s", strings.Join(e.Tags, ", ")))
	// }

	buf.WriteString(fmt.Sprintf("\n\nstarts at: %s", e.StartAt.Format(time.RFC1123Z)))

	if !e.EndAt.IsZero() {
		buf.WriteString(fmt.Sprintf("\nends at: %s", e.EndAt.Format(time.RFC1123Z)))
	}

	return buf.String()
}
