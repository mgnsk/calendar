package contract

import (
	"net/url"
	"time"

	"github.com/mgnsk/calendar/pkg/snowflake"
)

// EditEventForm is an edit event form.
type EditEventForm struct {
	EventID     snowflake.ID `in:"path=event_id"`
	IsDraft     bool         `in:"query=draft"`
	Title       string       `in:"form=title"`
	Description string       `in:"form=desc"`
	URL         string       `in:"form=url"`
	StartAt     string       `in:"form=start_at"`

	Location string `in:"form=location"`
	OSMType  string `in:"form=osm_type"`
	OSMID    uint64 `in:"form=osm_id"`

	Latitude     float64 `in:"form=latitude"`
	Longitude    float64 `in:"form=longitude"`
	UserTimezone string  `in:"form=user_timezone"`
}

// IsDraftOrNew reports whether the current event is draft or a new event.
func (r *EditEventForm) IsDraftOrNew() bool {
	return r.IsDraft || r.EventID == 0
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
