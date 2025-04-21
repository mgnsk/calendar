package contract

import (
	"net/url"
	"time"

	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/wreck"
)

// EditEventForm is an edit event form.
type EditEventForm struct {
	EventID        snowflake.ID `param:"event_id"`
	IsDraft        bool         `query:"draft"`
	Title          string       `form:"title"`
	Description    string       `form:"desc"`
	URL            string       `form:"url"`
	StartAt        string       `form:"start_at"`
	Location       string       `form:"location"`
	Latitude       float64      `form:"latitude"`
	Longitude      float64      `form:"longitude"`
	TimezoneOffset int          `form:"timezone_offset"`
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

	if r.StartAt == "" {
		errs.Set("start_at", "Required")
	}

	if r.Location == "" {
		errs.Set("location", "Required")
	}

	return errs
}

// ParseStartAt parses event start at time.
func (r *EditEventForm) ParseStartAt() (time.Time, error) {
	loc := time.FixedZone("", r.TimezoneOffset)

	startAt, err := time.ParseInLocation(FormDateTimeLayout, r.StartAt, loc)
	if err != nil {
		return time.Time{}, wreck.InvalidValue.New("Invalid start_at value", err)
	}

	return startAt, nil
}

// FormDateTimeLayout is the HTML datetime-local input time format.
const FormDateTimeLayout = "2006-01-02T15:04"
