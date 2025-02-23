package contract

import (
	"net/url"
	"time"

	"github.com/mgnsk/calendar/pkg/snowflake"
)

// EditEventForm is an edit event form.
type EditEventForm struct {
	EventID     snowflake.ID `param:"event_id"`
	Title       string       `form:"title"`
	Description string       `form:"desc"`
	URL         string       `form:"url"`
	StartAt     DateTime     `form:"start_at"`
	Location    string       `form:"location"`
}

// Validate the form.
func (r *EditEventForm) Validate() url.Values {
	errs := url.Values{}

	if r.Title == "" {
		errs.Set("title", "Required")
	}

	if r.Description == "" {
		errs.Set("desc", "Required")
	}

	if r.URL != "" {
		if _, err := url.Parse(r.URL); err != nil {
			errs.Set("url", "Invalid URL")
		}
	}

	if r.StartAt.value.IsZero() {
		errs.Set("start_at", "Required")
	}

	if r.Location == "" {
		errs.Set("location", "Required")
	}

	return errs
}

// NewDateTime creates a form datetime from time.Time.
func NewDateTime(ts time.Time) DateTime {
	return DateTime{value: ts}
}

// DateTime is a time type that parses from HTML datetime-local time format.
type DateTime struct {
	value time.Time
}

func (t *DateTime) String() string {
	if t.value.IsZero() {
		return ""
	}
	return t.value.Format(formDateTimeLayout)
}

// Time returns the time.Time value.
func (t *DateTime) Time() time.Time {
	return t.value
}

// UnmarshalText unmarshals the form datetime value.
// TODO: timezone from user
func (t *DateTime) UnmarshalText(text []byte) error {
	val := string(text)
	if val == "" {
		return nil
	}

	ts, err := time.Parse(formDateTimeLayout, val)
	if err != nil {
		return err
	}

	t.value = ts

	return nil
}

const formDateTimeLayout = "2006-01-02T15:04"
