package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
)

// Event is the event domain model.
type Event struct {
	ID           snowflake.ID
	StartAt      timestamp.Timestamp
	EndAt        timestamp.Timestamp
	Title        string
	Description  string
	URL          string
	Tags         []string
	TagRelations []*Tag
}

// GetCreatedAt returns the event created at time.
func (e *Event) GetCreatedAt() time.Time {
	return snowflake.ParseTime(e.ID.Int64())
}

func (e *Event) GetDateString() string {
	start := e.StartAt.Time()
	end := e.EndAt.Time()

	var buf strings.Builder
	buf.WriteString(start.Format("January _2, 2006 "))

	if start.Minute() == 0 {
		buf.WriteString(start.Format("3PM"))
	} else {
		buf.WriteString(start.Format("3:04PM"))
	}

	if !end.IsZero() {
		buf.WriteString("-")
		if end.Minute() == 0 {
			buf.WriteString(end.Format("3PM"))
		} else {
			buf.WriteString(end.Format("3:04PM"))
		}
	}

	return buf.String()

}

// GetDescription returns the event description with tags.
// TODO: test this
func (e *Event) GetDescription() string {
	var buf strings.Builder

	buf.WriteString(e.Description)

	if len(e.Tags) > 0 {
		buf.WriteString(fmt.Sprintf("\n\ntags: %s", strings.Join(e.Tags, ", ")))
	}

	buf.WriteString(fmt.Sprintf("\n\nstarts at: %s", e.StartAt.String()))

	if !e.EndAt.Time().IsZero() {
		buf.WriteString(fmt.Sprintf("\nends at: %s", e.EndAt.String()))
	}

	return buf.String()
}

// GetFTSData returns the combined data for full text search.
func (e *Event) GetFTSData() string {
	words := []string{
		e.Title,
		e.Description,
		e.URL,
		strings.Join(e.Tags, " "),
	}

	day := e.StartAt.Time().Day()
	words = append(words, timestamp.FormatDay(day))

	for _, format := range []string{
		"_2 January, 2006",
		"02.01.2006",
		time.DateOnly,
		"15:04",
		time.Kitchen,
		"3PM",
	} {
		words = append(words, e.StartAt.Time().Format(format))
	}

	return strings.Join(words, " ")
}
