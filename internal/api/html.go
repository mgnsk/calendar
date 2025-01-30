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
)

// HTMLHandler handles web pages.
type HTMLHandler struct {
	db *bun.DB
}

// Register the handler.
func (h *HTMLHandler) Register(e *echo.Echo) {
	e.GET(HomePath, h.Home)
}

// Home handles the home page.
func (h *HTMLHandler) Home(c echo.Context) error {
	// Lists events that started in the past 24 hours, start time ascending.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now().Add(-24*time.Hour), time.Time{}, "asc")
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	node := html.EventListPage(events)

	return node.Render(c.Response())
}

// NewHTMLHandler creates a new HTML handler.
func NewHTMLHandler(db *bun.DB) *HTMLHandler {
	return &HTMLHandler{
		db: db,
	}
}
