package handler

import (
	"fmt"
	"net/http"
	"strconv"

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
}

// Edit handles adding and editing events.
func (h *EditEventHandler) Edit(c echo.Context) error {
	user := loadUser(c)
	if user == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	s := loadSettings(c)
	csrf := c.Get("csrf").(string)

	req := contract.EditEventForm{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	var ev *domain.Event

	if req.EventID > 0 {
		event, err := model.GetEvent(c.Request().Context(), h.db, req.EventID)
		if err != nil {
			return err
		}

		if user.Role != domain.Admin && user.ID != event.UserID {
			return wreck.Forbidden.New("Non-admin users can only edit own events")
		}

		ev = event
	}

	switch c.Request().Method {
	case http.MethodGet:
		if ev != nil {
			req.Title = ev.Title
			req.Description = ev.Description
			req.URL = ev.URL
			req.StartAt = contract.NewDateTime(ev.StartAt)
			req.Location = ev.Location
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(req, nil, csrf)).Render(c.Response())

	case http.MethodPost:
		if errs := req.Validate(); len(errs) > 0 {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(req, errs, csrf)).Render(c.Response())
		}

		if ev != nil {
			ev.StartAt = req.StartAt.Time()
			ev.Title = req.Title
			ev.Description = req.Description
			ev.URL = req.URL
			ev.Location = req.Location
			// TODO: draft

			if err := model.UpdateEvent(c.Request().Context(), h.db, ev); err != nil {
				return err
			}

			// TODO: add success flash message
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/edit/%d", ev.ID))
		}

		eventID := snowflake.Generate()

		if err := model.InsertEvent(c.Request().Context(), h.db, &domain.Event{
			ID:          eventID,
			StartAt:     req.StartAt.Time(),
			Title:       req.Title,
			Description: req.Description,
			URL:         req.URL,
			Location:    req.Location,
			IsDraft:     false, // TODO
			UserID:      user.ID,
		}); err != nil {
			return err
		}

		// TODO: add success flash message
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/edit/%d", eventID))

	default:
		return wreck.NotFound.New("Not found")
	}
}

// Delete handles deleting events.
func (h *EditEventHandler) Delete(c echo.Context) error {
	user := loadUser(c)
	if user == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	eventID := c.Param("event_id")
	if eventID == "" {
		return wreck.InvalidValue.New("Expected event_id path param")
	}

	id, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		return wreck.InvalidValue.New("Invalid event_id path param", err)
	}

	ev, err := model.GetEvent(c.Request().Context(), h.db, snowflake.ID(id))
	if err != nil {
		return err
	}

	if user.Role != domain.Admin && user.ID != ev.UserID {
		return wreck.Forbidden.New("Non-admin users can only edit own events")
	}

	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		if err := model.DeleteEvent(c.Request().Context(), h.db, ev); err != nil {
			return err
		}

		// TODO: add success flash message
		hxhttp.SetRefresh(c.Response().Header())

		return nil
	}

	return wreck.NotFound.New("Not found")
}

// Preview returns a preview of the event.
func (h *EditEventHandler) Preview(c echo.Context) error {
	user := loadUser(c)
	if user == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	req := contract.EditEventForm{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	ev := &domain.Event{
		StartAt:     req.StartAt.Time(),
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Location:    req.Location,
		IsDraft:     false, // TODO
	}

	csrf := c.Get("csrf").(string)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)
	return html.EventCard(nil, ev, csrf).Render(c.Response())
}

// Register the handler.
func (h *EditEventHandler) Register(g *echo.Group) {
	g.GET("/edit/:event_id", h.Edit)
	g.POST("/edit/:event_id", h.Edit)

	g.POST("/delete/:event_id", h.Delete)

	g.POST("/preview", h.Preview)
}

// NewEditEventHandler creates a new edit event handler.
func NewEditEventHandler(db *bun.DB) *EditEventHandler {
	return &EditEventHandler{
		db: db,
	}
}
