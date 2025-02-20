package handler

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/uptrace/bun"
	"github.com/yuin/goldmark"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EditEventHandler handles adding and editing events.
type EditEventHandler struct {
	db *bun.DB
}

// Add event handles adding events.
func (h *EditEventHandler) Add(c echo.Context) error {
	user := loadUser(c)
	if user == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	s := loadSettings(c)
	csrf := c.Get("csrf").(string)

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(nil, nil, csrf)).Render(c.Response())

	case http.MethodPost:
		form, err := c.FormParams()
		if err != nil {
			return err
		}
		errs := url.Values{}

		// TODO: form validation framework
		title := strings.TrimSpace(c.FormValue("title"))
		desc := strings.TrimSpace(c.FormValue("desc"))

		if title == "" || desc == "" {
			errs.Set("title", "Required")
			errs.Set("desc", "Required")

			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, errs, csrf)).Render(c.Response())
		}

		if err := goldmark.Convert([]byte(desc), io.Discard); err != nil {
			errs.Set("desc", "Invalid markdown")

			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, errs, csrf)).Render(c.Response())
		}

		if err := model.InsertEvent(c.Request().Context(), h.db, &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(2 * time.Hour),
			Title:       title,
			Description: desc,
			URL:         "",
			IsDraft:     false, // TODO
			UserID:      user.ID,
		}); err != nil {
			return err
		}

		// TODO: add success flash message
		return c.Redirect(http.StatusSeeOther, "/")

	default:
		panic("unhandled method")
	}
}

// Edit handles editing events.
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

	ev, err := model.GetEvent(c.Request().Context(), h.db, snowflake.ID(id))
	if err != nil {
		return err
	}

	if user.Role != domain.Admin && user.ID != ev.UserID {
		return wreck.Forbidden.New("Non-admin users can only edit own events")
	}

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		form := url.Values{}
		form.Set("title", ev.Title)
		form.Set("desc", ev.Description)

		return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, nil, csrf)).Render(c.Response())

	case http.MethodPost:
		form, err := c.FormParams()
		if err != nil {
			return err
		}
		errs := url.Values{}

		title := strings.TrimSpace(c.FormValue("title"))
		desc := strings.TrimSpace(c.FormValue("desc"))

		if title == "" || desc == "" {
			errs.Set("title", "Required")
			errs.Set("desc", "Required")

			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, errs, csrf)).Render(c.Response())
		}

		if err := goldmark.Convert([]byte(desc), io.Discard); err != nil {
			errs.Set("desc", "Invalid markdown")

			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.EditEventMain(form, errs, csrf)).Render(c.Response())
		}

		ev.Title = title
		ev.Description = desc

		if err := model.UpdateEvent(c.Request().Context(), h.db, ev); err != nil {
			return err
		}

		// TODO: add success flash message
		return c.Redirect(http.StatusSeeOther, "/")

	default:
		panic("unhandled method")
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

	title := strings.TrimSpace(c.FormValue("title"))
	desc := strings.TrimSpace(c.FormValue("desc"))

	if err := goldmark.Convert([]byte(desc), io.Discard); err != nil {
		return wreck.InvalidValue.New("Invalid markdown", err)
	}

	ev := &domain.Event{
		ID:          0,
		StartAt:     time.Time{}, // TODO
		EndAt:       time.Time{}, // TODO
		Title:       title,
		Description: desc,
		URL:         "", // TODO
	}

	csrf := c.Get("csrf").(string)

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)
	return html.EventCard(nil, ev, csrf).Render(c.Response())
}

// Register the handler.
func (h *EditEventHandler) Register(g *echo.Group) {
	g.GET("/add", h.Add)
	g.POST("/add", h.Add)

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
