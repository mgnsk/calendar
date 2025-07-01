package contract

import (
	"net/url"
	"time"

	"github.com/mgnsk/calendar/pkg/snowflake"
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
	UserTimezone   string       `form:"user_timezone"`
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
	} else if _, err := time.Parse(FormDateTimeLayout, r.StartAt); err != nil {
		errs.Set("start_at", "Invalid format")
	}

	if r.Location == "" {
		errs.Set("location", "Required")
	}

	return errs
}

// FormDateTimeLayout is the HTML datetime-local input time format.
const FormDateTimeLayout = "2006-01-02T15:04"
