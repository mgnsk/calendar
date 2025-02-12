package handler

import (
	"net/http"
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

	settings := loadSettings(c)

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.AddEventPage(settings.Title, nil, "", c.Get("csrf").(string)).Render(c.Response())

	case http.MethodPost:
		errs := map[string]string{}

		// TODO: form validation framework
		title := c.FormValue("title")
		if title == "" {
			errs["title"] = "Required"

			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.AddEventPage(settings.Title, errs, title, c.Get("csrf").(string)).Render(c.Response())
		}

		if err := model.InsertEvent(c.Request().Context(), h.db, &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now(),
			EndAt:       time.Now().Add(2 * time.Hour),
			Title:       title,
			Description: "",
			URL:         "",
			Tags:        []string{"festival", "rock"},
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
