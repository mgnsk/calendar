package api

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/uptrace/bun"
)

// Paths for web endpoints.
const (
	HomePath = "/"
	PastPath = "/past"
)

// HTMLHandler handles web pages.
type HTMLHandler struct {
	db *bun.DB
}

// HTMLMiddleware sets headers for HTML responses.
func HTMLMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

		return next(c)
	}
}

// Register the handler.
func (h *HTMLHandler) Register(e *echo.Echo) {
	g := e.Group("", HTMLMiddleware)
	g.GET(HomePath, h.Home)
	g.GET(PastPath, h.PastEvents)
}

// Home handles the home page.
func (h *HTMLHandler) Home(c echo.Context) error {
	// Lists events that started in the past 24 hours, start time ascending.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now().Add(-24*time.Hour), time.Time{}, "asc")
	if err != nil {
		return err
	}

	c.Response().WriteHeader(200)

	node := html.CurrentEventsPage(events)

	return node.Render(c.Response())
}

// PastEvents handles past events page.
func (h *HTMLHandler) PastEvents(c echo.Context) error {
	// Lists events that have already started, in descending order.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Time{}, time.Now(), "desc")
	if err != nil {
		return err
	}

	c.Response().WriteHeader(200)

	node := html.PastEventsPage(events)

	return node.Render(c.Response())
}

// NewHTMLHandler creates a new HTML handler.
func NewHTMLHandler(db *bun.DB) *HTMLHandler {
	return &HTMLHandler{
		db: db,
	}
}
