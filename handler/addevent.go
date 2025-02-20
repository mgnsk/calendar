package handler

import (
	"io"
	"net/http"
	"net/url"
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
)

// AddEventHandler handles adding events.
type AddEventHandler struct {
	db *bun.DB
}

// Add event handles adding events.
func (h *AddEventHandler) Add(c echo.Context) error {
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

		return html.Page(s.Title, user, c.Path(), csrf, html.AddEventMain(nil, nil, csrf)).Render(c.Response())

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

			return html.Page(s.Title, user, c.Path(), csrf, html.AddEventMain(form, errs, csrf)).Render(c.Response())
		}

		if err := goldmark.Convert([]byte(desc), io.Discard); err != nil {
			errs.Set("desc", "Invalid markdown")

			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, user, c.Path(), csrf, html.AddEventMain(form, errs, csrf)).Render(c.Response())
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

// Preview returns a preview of the event.
func (h *AddEventHandler) Preview(c echo.Context) error {
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

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)
	return html.EventCard(ev).Render(c.Response())
}

// Register the handler.
func (h *AddEventHandler) Register(g *echo.Group) {
	g.GET("/add", h.Add)
	g.POST("/add", h.Add)

	g.POST("/preview", h.Preview)
}

// NewAddEventHandler creates a new add event handler.
func NewAddEventHandler(db *bun.DB) *AddEventHandler {
	return &AddEventHandler{
		db: db,
	}
}
