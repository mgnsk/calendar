package handler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
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
		title := c.FormValue("title")
		desc := c.FormValue("desc")
		if title == "" || desc == "" {
			errs.Set("title", "Required")
			errs.Set("desc", "Required")

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
		}); err != nil {
			return err
		}

		// TODO: add success flash message
		return c.Redirect(http.StatusSeeOther, "/")

	default:
		panic("unhandled method")
	}
}

// Register the handler.
func (h *AddEventHandler) Register(g *echo.Group) {
	g.GET("/add", h.Add)
	g.POST("/add", h.Add)
}

// NewAddEventHandler creates a new add event handler.
func NewAddEventHandler(db *bun.DB) *AddEventHandler {
	return &AddEventHandler{
		db: db,
	}
}
