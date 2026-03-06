package handler

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EventsHandler handles event pages rendering.
type EventsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Upcoming handles upcoming events.
func (h *EventsHandler) Upcoming(c *server.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtFrom(time.Now()),
		model.OrderStartAtAsc,
	)
}

// Past handles past events.
func (h *EventsHandler) Past(c *server.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtUntil(time.Now()),
		model.OrderStartAtDesc,
	)
}

// MyEvents handles current user events.
func (h *EventsHandler) MyEvents(c *server.Context) error {
	if c.User == nil {
		return calendar.Forbidden.New("Must be logged in")
	}

	return h.events(
		c,
		model.NewEventsQuery().WithUserID(c.User.ID).WithIncludeDrafts(),
		model.OrderCreatedAtDesc,
	)
}

// Tags handles tags.
func (h *EventsHandler) Tags(c *server.Context) error {
	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		tags, err := model.ListTags(c.Request().Context(), h.db, time.Now(), 500)
		if err != nil {
			if !errors.Is(err, calendar.NotFound) {
				return err
			}
		}

		slices.SortFunc(tags, func(a, b *domain.Tag) int {
			return strings.Compare(a.Name, b.Name)
		})

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.TagListPartial(tags, c.CSRF).Render(c.Response())
	}

	return server.RenderPage(c, h.sm,
		html.TagsMain(c.CSRF),
	)
}

func (h *EventsHandler) events(c *server.Context, query model.EventsQueryBuilder, order model.EventOrder) error {
	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		req := contract.ListEventsRequest{}
		if err := c.Bind(&req); err != nil {
			return err
		}

		var cursor int64

		switch c.Path() {
		case "/my-events":
			cursor = req.LastID

		case "/", "/past":
			cursor = req.Offset

		default:
			return calendar.NotFound.New("Not found")
		}

		query = query.
			WithOrder(cursor, order).
			WithLimit(contract.EventLimitPerPage).
			WithSearchText(req.Search)

		var (
			events []*domain.Event
			err    error
		)

		events, err = query.List(c.Request().Context(), h.db)
		if err != nil {
			if !errors.Is(err, calendar.NotFound) {
				return err
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(c.User, cursor, events, c.CSRF).Render(c.Response())
	}

	return server.RenderPage(c, h.sm,
		html.EventsMain(c.CSRF),
	)
}

// Register the handler.
func (h *EventsHandler) Register(g *echo.Group) {
	g.GET("/", server.Wrap(h.db, h.sm, h.Upcoming))
	g.POST("/", server.Wrap(h.db, h.sm, h.Upcoming)) // For htmx.

	g.GET("/past", server.Wrap(h.db, h.sm, h.Past))
	g.POST("/past", server.Wrap(h.db, h.sm, h.Past)) // For htmx.

	g.GET("/tags", server.Wrap(h.db, h.sm, h.Tags))
	g.POST("/tags", server.Wrap(h.db, h.sm, h.Tags)) // For htmx.

	g.GET("/my-events", server.Wrap(h.db, h.sm, h.MyEvents))
	g.POST("/my-events", server.Wrap(h.db, h.sm, h.MyEvents)) // For htmx.
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(db *bun.DB, sm *scs.SessionManager) *EventsHandler {
	return &EventsHandler{
		db: db,
		sm: sm,
	}
}
