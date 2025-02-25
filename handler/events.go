package handler

import (
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EventLimitPerPage specifies maximum number of events per page.
const EventLimitPerPage = 25

// EventsHandler handles event pages rendering.
type EventsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Latest handles latest events.
func (h *EventsHandler) Latest(c echo.Context) error {
	return h.events(
		c,
		model.NewEventsQuery(),
		model.OrderCreatedAtDesc,
	)
}

// Upcoming handles upcoming events.
func (h *EventsHandler) Upcoming(c echo.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtFrom(time.Now()),
		model.OrderStartAtAsc,
	)
}

// Past handles past events.
func (h *EventsHandler) Past(c echo.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtUntil(time.Now()),
		model.OrderStartAtDesc,
	)
}

// MyEvents handles current user events.
func (h *EventsHandler) MyEvents(c echo.Context) error {
	rc := GetContext(c)
	if rc.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	return h.events(
		c,
		model.NewEventsQuery().WithUserID(rc.User.ID),
		model.OrderCreatedAtDesc,
	)
}

// Tags handles tags.
func (h *EventsHandler) Tags(c echo.Context) error {
	rc := GetContext(c)

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

		return html.TagListPartial(tags, rc.CSRF).Render(c.Response())
	}

	return RenderPage(c, rc,
		html.TagsMain(rc.CSRF),
	)
}

func (h *EventsHandler) events(c echo.Context, query model.EventsQueryBuilder, order model.EventOrder) error {
	rc := GetContext(c)

	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		var cursor int64

		if strings.HasPrefix(c.Path(), "/upcoming") || strings.HasPrefix(c.Path(), "/past") {
			if offset := getIntForm("offset", c); offset != nil {
				cursor = *offset + EventLimitPerPage
			}
		} else {
			if lastID := getIntForm("last_id", c); lastID != nil {
				cursor = *lastID
			}
		}

		query = query.
			WithOrder(cursor, order).
			WithLimit(EventLimitPerPage)

		var (
			events []*domain.Event
			err    error
		)

		events, err = query.List(c.Request().Context(), h.db, false, c.FormValue("search"))
		if err != nil {
			if !errors.Is(err, wreck.NotFound) {
				return err
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(rc.User, cursor, events, rc.CSRF).Render(c.Response())
	}

	return RenderPage(c, rc,
		html.EventsMain(rc.CSRF),
	)
}

// Register the handler.
func (h *EventsHandler) Register(g *echo.Group) {
	g.GET("/", h.Latest)
	g.POST("/", h.Latest) // Fox htmx.

	g.GET("/upcoming", h.Upcoming)
	g.POST("/upcoming", h.Upcoming) // Fox htmx.

	g.GET("/past", h.Past)
	g.POST("/past", h.Past) // For htmx.

	g.GET("/tags", h.Tags)
	g.POST("/tags", h.Tags) // Fox htmx.

	g.GET("/my-events", h.MyEvents)
	g.POST("/my-events", h.MyEvents) // Fox htmx.
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(db *bun.DB, sm *scs.SessionManager) *EventsHandler {
	return &EventsHandler{
		db: db,
		sm: sm,
	}
}

func getIntForm(key string, c echo.Context) *int64 {
	if val := c.FormValue(key); val != "" {
		v, _ := strconv.ParseInt(c.FormValue(key), 10, 64)
		return &v
	}
	return nil
}
