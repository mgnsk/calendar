package handler

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EventsHandler handles event pages rendering.
type EventsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Latest handles latest events.
func (h *EventsHandler) Latest(c *Context) error {
	return h.events(
		c,
		model.NewEventsQuery(),
		model.OrderCreatedAtDesc,
	)
}

// Upcoming handles upcoming events.
func (h *EventsHandler) Upcoming(c *Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtFrom(time.Now()),
		model.OrderStartAtAsc,
	)
}

// Past handles past events.
func (h *EventsHandler) Past(c *Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtUntil(time.Now()),
		model.OrderStartAtDesc,
	)
}

// MyEvents handles current user events.
func (h *EventsHandler) MyEvents(c *Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	return h.events(
		c,
		model.NewEventsQuery().WithUserID(c.User.ID),
		model.OrderCreatedAtDesc,
	)
}

// Tags handles tags.
func (h *EventsHandler) Tags(c *Context) error {
	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		tags, err := model.ListTags(c.Request().Context(), h.db, 500)
		if err != nil {
			if !errors.Is(err, wreck.NotFound) {
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

	return RenderPage(c, h.sm,
		html.TagsMain(c.CSRF),
	)
}

func (h *EventsHandler) events(c *Context, query model.EventsQueryBuilder, order model.EventOrder) error {
	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		req := contract.ListEventsRequest{}

		if err := c.Bind(&req); err != nil {
			return err
		}

		var cursor int64

		switch c.Path() {
		case "/", "/my-events":
			cursor = req.LastID

		case "/upcoming", "/past":
			cursor = req.Offset

		default:
			return wreck.NotFound.New("Not found")
		}

		query = query.
			WithOrder(cursor, order).
			WithLimit(contract.EventLimitPerPage)

		var (
			events []*domain.Event
			err    error
		)

		events, err = query.List(c.Request().Context(), h.db, false, req.Search)
		if err != nil {
			if !errors.Is(err, wreck.NotFound) {
				return err
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(c.User, cursor, events, c.CSRF).Render(c.Response())
	}

	return RenderPage(c, h.sm,
		html.EventsMain(c.CSRF),
	)
}

// Register the handler.
func (h *EventsHandler) Register(g *echo.Group) {
	g.GET("/", Wrap(h.db, h.sm, h.Latest))
	g.POST("/", Wrap(h.db, h.sm, h.Latest)) // Fox htmx.

	g.GET("/upcoming", Wrap(h.db, h.sm, h.Upcoming))
	g.POST("/upcoming", Wrap(h.db, h.sm, h.Upcoming)) // Fox htmx.

	g.GET("/past", Wrap(h.db, h.sm, h.Past))
	g.POST("/past", Wrap(h.db, h.sm, h.Past)) // For htmx.

	g.GET("/tags", Wrap(h.db, h.sm, h.Tags))
	g.POST("/tags", Wrap(h.db, h.sm, h.Tags)) // Fox htmx.

	g.GET("/my-events", Wrap(h.db, h.sm, h.MyEvents))
	g.POST("/my-events", Wrap(h.db, h.sm, h.MyEvents)) // Fox htmx.
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(db *bun.DB, sm *scs.SessionManager) *EventsHandler {
	return &EventsHandler{
		db: db,
		sm: sm,
	}
}
