package contract

import "github.com/mgnsk/calendar/pkg/snowflake"

// ListEventsRequest is a request to list events.
type ListEventsRequest struct {
	Offset int64  `in:"query=offset"`
	LastID int64  `in:"query=last_id"`
	Search string `in:"query=search"`
}

// DeleteEventRequest is a request to delete an event.
type DeleteEventRequest struct {
	EventID snowflake.ID `in:"path=event_id"`
}

// EventLimitPerPage specifies maximum number of events per page.
const EventLimitPerPage = 25
