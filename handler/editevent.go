package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/markdown"
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

	eventID := c.Param("event_id")
	if eventID == "" {
		return wreck.InvalidValue.New("Expected event_id path param")
	}

	id, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		return wreck.InvalidValue.New("Invalid event_id path param", err)
	}

	var ev *domain.Event

	if id == 0 {
		ev = &domain.Event{}
	} else {
		event, err := model.GetEvent(c.Request().Context(), h.db, snowflake.ID(id))
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
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		form := url.Values{}
		form.Set("title", ev.Title)
		form.Set("desc", ev.Description)
		form.Set("url", ev.URL)
		if !ev.StartAt.IsZero() {
			form.Set("start_at", ev.StartAt.Format(html.DateTimeFormat))
		}

		return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, nil, ev.ID, csrf)).Render(c.Response())

	case http.MethodPost:
		form, err := c.FormParams()
		if err != nil {
			return err
		}

		data, errs := parseEvent(c)
		if len(errs) > 0 {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, errs, ev.ID, csrf)).Render(c.Response())
		}

		if ev.ID == 0 {
			ev.ID = snowflake.Generate()

			if err := model.InsertEvent(c.Request().Context(), h.db, &domain.Event{
				ID:          ev.ID,
				StartAt:     data.StartAt,
				Title:       data.Title,
				Description: data.Description,
				URL:         data.URL,
				IsDraft:     data.IsDraft,
				UserID:      user.ID,
			}); err != nil {
				return err
			}
		} else {
			ev.Title = data.Title
			ev.Description = data.Description
			ev.URL = data.URL
			ev.StartAt = data.StartAt

			if err := model.UpdateEvent(c.Request().Context(), h.db, ev); err != nil {
				return err
			}
		}

		// TODO: add success flash message
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/edit/%d", ev.ID))

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

	ev, errs := parseEvent(c)
	if errs.Has("description") {
		// TODO: dependency on parseEvent internals, see comment there.
		return wreck.InvalidValue.New("Invalid markdown")
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

// parseEvent parses an event from form input.
// TODO: consider defining new types (EditEventRequest?) and using
// some form binding library instead of manually using domain.Event here.
func parseEvent(c echo.Context) (*domain.Event, url.Values) {
	title := strings.TrimSpace(c.FormValue("title"))
	desc := strings.TrimSpace(c.FormValue("desc"))
	eventURL := strings.TrimSpace(c.FormValue("url"))
	startAtVal := strings.TrimSpace(c.FormValue("start_at"))

	errs := url.Values{}

	// TODO: improve form validation
	if title == "" || desc == "" || eventURL == "" || startAtVal == "" {
		errs.Set("title", "Required")
		errs.Set("desc", "Required")
		errs.Set("url", "Required")
		errs.Set("start_at", "Required")
	} else if _, err := markdown.Convert(desc); err != nil {
		errs.Set("desc", "Invalid markdown")
		// TODO: refactor, the preview handler
		// needs to preview as much as possible partially valid event.
		return nil, errs
	}

	u, err := url.Parse(eventURL)
	if err != nil {
		errs.Set("url", "Invalid URL")
	}

	startAt, err := time.Parse(html.DateTimeFormat, startAtVal)
	if err != nil {
		errs.Set("start_at", "Invalid start at datetime")
	}

	ev := &domain.Event{
		ID:          0,
		StartAt:     startAt,
		Title:       title,
		Description: desc,
		URL:         u.String(),
		IsDraft:     false, // TODO
		UserID:      0,     // TODO: not used
	}

	return ev, errs
}
