package api

import (
	"net/http"

	ics "github.com/arran4/golang-ical"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/uptrace/bun"
)

// ICalHandler handles iCal output.
type ICalHandler struct {
	db     *bun.DB
	config FeedConfig
}

// Handle the iCal endpoint.
func (h *ICalHandler) Handle(c echo.Context) error {
	events, err := model.ListEvents(c.Request().Context(), h.db, "asc")
	if err != nil {
		return err
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription(h.config.Title)
	cal.SetUrl(h.config.Link) // TODO: is this the calendar URL or general URL?

	for _, ev := range events {
		event := cal.AddEvent(ev.ID.String())

		event.SetCreatedTime(ev.GetCreatedAt())
		event.SetModifiedAt(ev.GetCreatedAt())

		event.SetStartAt(ev.StartAt.Time())
		if endAt := ev.EndAt.Time(); !endAt.IsZero() {
			event.SetEndAt(endAt)
		}

		event.SetSummary(ev.Title)
		event.SetURL(ev.URL)
		event.SetDescription(ev.GetDescription())
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/calendar; charset=utf-8")
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="calendar.ics"`)
	c.Response().WriteHeader(http.StatusOK)

	return cal.SerializeTo(c.Response())
}

// NewICalHandler creates a new iCal handler.
func NewICalHandler(db *bun.DB, config FeedConfig) *ICalHandler {
	return &ICalHandler{
		db:     db,
		config: config,
	}
}
