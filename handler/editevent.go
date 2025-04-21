package handler

import (
	"context"
	"encoding/json"
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
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EditEventHandler handles adding and editing events.
type EditEventHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Edit handles adding and editing events.
func (h *EditEventHandler) Edit(c *Context) error {
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

		return RenderPage(c, h.sm,
			html.EditEventMain(req, nil, c.CSRF),
		)

	case http.MethodPost:
		if errs := req.Validate(); len(errs) > 0 {
			return RenderPage(c, h.sm,
				html.EditEventMain(req, errs, c.CSRF),
			)
		}

		startAt, err := req.ParseStartAt()
		if err != nil {
			return err
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
func (h *EditEventHandler) Delete(c *Context) error {
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
func (h *EditEventHandler) Preview(c *Context) error {
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

	startAt, err := req.ParseStartAt()
	if err != nil {
		return err
	}

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

// GetTimezone returns timezone offset by geo coordinates.
func (h *EditEventHandler) GetTimezone(c *Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	req := contract.GetTimezoneRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	u, err := url.Parse("https://api.geotimezone.com/public/timezone")
	if err != nil {
		panic(err)
	}

	q := url.Values{}
	q.Set("latitude", req.Latitude)
	q.Set("longitude", req.Longitude)

	u.RawQuery = q.Encode()

	r, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	r = r.WithContext(ctx)

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return wreck.Internal.New(fmt.Sprintf("Unable to reach geotimezone API: status %d", res.StatusCode))
	}

	var timezoneResponse struct {
		IANATimezone string `json:"iana_timezone"`
	}

	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&timezoneResponse); err != nil {
		return err
	}

	ianaTimezone := timezoneResponse.IANATimezone

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
			return wreck.InvalidValue.New("Invalid location timezone", err)
		}

		loc = l
	}

	_, offset := time.Now().In(loc).Zone()

	return c.JSON(http.StatusOK, contract.GetTimezoneResponse{
		TimezoneOffset: offset,
	})
}

// Register the handler.
func (h *EditEventHandler) Register(g *echo.Group) {
	g.GET("/edit/:event_id", Wrap(h.db, h.sm, h.Edit))
	g.POST("/edit/:event_id", Wrap(h.db, h.sm, h.Edit))

	g.POST("/delete/:event_id", Wrap(h.db, h.sm, h.Delete))

	g.POST("/preview", Wrap(h.db, h.sm, h.Preview))
	g.GET("/gettimezone", Wrap(h.db, h.sm, h.GetTimezone))
}

// NewEditEventHandler creates a new edit event handler.
func NewEditEventHandler(db *bun.DB, sm *scs.SessionManager) *EditEventHandler {
	return &EditEventHandler{
		db: db,
		sm: sm,
	}
}
