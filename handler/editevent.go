package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// TimezoneFinder finds timezone by geo coordinates.
type TimezoneFinder interface {
	GetTimezoneName(lng, lat float64) string
}

// EditEventHandler handles adding and editing events.
type EditEventHandler struct {
	db     *bun.DB
	sm     *scs.SessionManager
	finder TimezoneFinder
}

// Edit handles adding and editing events.
func (h *EditEventHandler) Edit(c *server.Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	req := contract.EditEventForm{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := (&echo.DefaultBinder{}).BindQueryParams(c, &req); err != nil {
		return err
	}

	var ev *domain.Event

	if req.EventID > 0 {
		event, err := model.GetEvent(c.Request().Context(), h.db, req.EventID)
		if err != nil {
			return err
		}

		if c.User.Role != domain.Admin && c.User.ID != event.UserID {
			return wreck.Forbidden.New("Non-admin users can only edit own events")
		}

		ev = event
	}

	switch c.Request().Method {
	case http.MethodGet:
		if ev != nil {
			req.Title = ev.Title
			req.IsDraft = ev.IsDraft
			req.Description = ev.Description
			req.URL = ev.URL
			req.StartAt = ev.StartAt.Format(contract.FormDateTimeLayout)
			req.Location = ev.Location
			req.Latitude = ev.Latitude
			req.Longitude = ev.Longitude
			_, offset := ev.StartAt.Zone()
			req.TimezoneOffset = offset
		}

		return server.RenderPage(c, h.sm,
			html.EditEventMain(req, nil, c.CSRF),
		)

	case http.MethodPost:
		if errs := req.Validate(); len(errs) > 0 {
			return server.RenderPage(c, h.sm,
				html.EditEventMain(req, errs, c.CSRF),
			)
		}

		startAt, err := h.parseStartAt(req)
		if err != nil {
			errs := url.Values{}
			errs.Set("start_at", "Invalid start_at value")
			return server.RenderPage(c, h.sm,
				html.EditEventMain(req, errs, c.CSRF),
			)
		}

		if ev != nil {
			ev.StartAt = startAt
			ev.Title = req.Title
			ev.IsDraft = req.IsDraft
			ev.Description = req.Description
			ev.URL = req.URL
			ev.Location = req.Location
			ev.Latitude = req.Latitude
			ev.Longitude = req.Longitude

			if err := model.UpdateEvent(c.Request().Context(), h.db, ev); err != nil {
				return err
			}

			if req.IsDraft {
				h.sm.Put(c.Request().Context(), "flash-success", "Draft saved")
			} else {
				h.sm.Put(c.Request().Context(), "flash-success", "Event published")
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/edit/%d", ev.ID))
		}

		eventID := snowflake.Generate()

		if err := model.InsertEvent(c.Request().Context(), h.db, &domain.Event{
			ID:          eventID,
			StartAt:     startAt,
			Title:       req.Title,
			Description: req.Description,
			URL:         req.URL,
			Location:    req.Location,
			Latitude:    req.Latitude,
			Longitude:   req.Longitude,
			IsDraft:     req.IsDraft,
			UserID:      c.User.ID,
		}); err != nil {
			return err
		}

		if req.IsDraft {
			h.sm.Put(c.Request().Context(), "flash-success", "Draft saved")
		} else {
			h.sm.Put(c.Request().Context(), "flash-success", "Event published")
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/edit/%d", eventID))

	default:
		return wreck.NotFound.New("Not found")
	}
}

// Delete handles deleting events.
func (h *EditEventHandler) Delete(c *server.Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	req := contract.DeleteEventRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	ev, err := model.GetEvent(c.Request().Context(), h.db, req.EventID)
	if err != nil {
		return err
	}

	if c.User.Role != domain.Admin && c.User.ID != ev.UserID {
		return wreck.Forbidden.New("Non-admin users can only edit own events")
	}

	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		if err := model.DeleteEvent(c.Request().Context(), h.db, ev); err != nil {
			return err
		}

		h.sm.Put(c.Request().Context(), "flash-success", "Event deleted")

		hxhttp.SetRefresh(c.Response().Header())

		return nil
	}

	return wreck.NotFound.New("Not found")
}

// Preview returns a preview of the event.
func (h *EditEventHandler) Preview(c *server.Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	req := contract.EditEventForm{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := (&echo.DefaultBinder{}).BindQueryParams(c, &req); err != nil {
		return err
	}

	startAt, _ := h.parseStartAt(req)

	ev := &domain.Event{
		StartAt:     startAt,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Location:    req.Location,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsDraft:     req.IsDraft,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)
	return html.EventCard(nil, ev, c.CSRF).Render(c.Response())
}

// Register the handler.
func (h *EditEventHandler) Register(g *echo.Group) {
	g.GET("/edit/:event_id", server.Wrap(h.db, h.sm, h.Edit))
	g.POST("/edit/:event_id", server.Wrap(h.db, h.sm, h.Edit))

	g.POST("/delete/:event_id", server.Wrap(h.db, h.sm, h.Delete))

	g.POST("/preview", server.Wrap(h.db, h.sm, h.Preview))
}

func (h *EditEventHandler) parseStartAt(req contract.EditEventForm) (time.Time, error) {
	ianaTimezone := h.finder.GetTimezoneName(req.Longitude, req.Latitude)

	if ianaTimezone == "" {
		// If timezone not found, fall back to user timezone.
		ianaTimezone = req.UserTimezone
	}

	var loc *time.Location

	if ianaTimezone == "" {
		// if user timezone also not found, fall back to UTC.
		loc = time.UTC
	} else {
		l, err := time.LoadLocation(ianaTimezone)
		if err != nil {
			// TODO: should we log this error and still use UTC?
			return time.Time{}, wreck.InvalidValue.New("Invalid location timezone", err)
		}

		loc = l
	}

	startAt, err := time.ParseInLocation(contract.FormDateTimeLayout, req.StartAt, loc)
	if err != nil {
		return time.Time{}, wreck.InvalidValue.New("Invalid start_at value", err)
	}

	return startAt, nil
}

// NewEditEventHandler creates a new edit event handler.
func NewEditEventHandler(db *bun.DB, sm *scs.SessionManager, finder TimezoneFinder) *EditEventHandler {
	return &EditEventHandler{
		db:     db,
		sm:     sm,
		finder: finder,
	}
}
