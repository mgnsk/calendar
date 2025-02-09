package handler

import (
	"cmp"
	"errors"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/mgnsk/evcache/v4"
	slogecho "github.com/samber/slog-echo"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EventLimitPerPage specifies maximum number of events per page.
const EventLimitPerPage = 3

// EventsHandler handles event pages rendering.
type EventsHandler struct {
	db          *bun.DB
	tagsCache   *evcache.Cache[string, []*domain.Tag]
	eventsCache *evcache.Cache[string, []*domain.Event]
}

// Latest handles latest events.
func (h *EventsHandler) Latest(c echo.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtFrom(time.Now()),
		model.OrderCreatedAtDesc,
		"Latest Events",
	)
}

// Upcoming handles upcoming events.
func (h *EventsHandler) Upcoming(c echo.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtFrom(time.Now()),
		model.OrderStartAtAsc,
		"Upcoming Events",
	)
}

// Past handles past events.
func (h *EventsHandler) Past(c echo.Context) error {
	return h.events(
		c,
		model.NewEventsQuery().WithStartAtUntil(time.Now()),
		model.OrderStartAtDesc,
		"Past Events",
	)
}

// Tags handles tags.
func (h *EventsHandler) Tags(c echo.Context) error {
	user := loadUser(c)

	if hxhttp.IsRequest(c.Request().Header) {
		var (
			tags []*domain.Tag
			err  error
		)

		if user != nil {
			slogecho.AddCustomAttributes(c, slog.Bool("cached", false))
			tags, err = model.ListTags(c.Request().Context(), h.db)
		} else {
			didFetch := false
			tags, err = h.tagsCache.Fetch(c.Request().URL.String(), func() ([]*domain.Tag, error) {
				didFetch = true
				return model.ListTags(c.Request().Context(), h.db)
			})
			slogecho.AddCustomAttributes(c, slog.Bool("cached", !didFetch))
		}

		if err != nil {
			if !errors.Is(err, wreck.NotFound) {
				return err
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.TagListPartial(tags).Render(c.Response())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	settings := loadSettings(c)

	return html.TagsPage(html.TagsPageParams{
		MainTitle:    settings.Title,
		SectionTitle: "Tags",
		Path:         c.Path(),
		User:         user,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

func (h *EventsHandler) events(c echo.Context, query model.EventsQueryBuilder, order model.EventOrder, sectionTitle string) error {
	user := loadUser(c)

	filterTag := getTagFilter(c)

	if hxhttp.IsRequest(c.Request().Header) {
		lastID := getIntQuery("last_id", c)

		offset := getIntQuery("offset", c)
		if offset > 0 {
			offset += EventLimitPerPage
		}

		cursor := cmp.Or(offset, lastID)

		query = query.
			WithOrder(cursor, order).
			WithFilterTags(filterTag).
			WithLimit(EventLimitPerPage)

		var (
			events []*domain.Event
			err    error
		)

		if user != nil {
			slogecho.AddCustomAttributes(c, slog.Bool("cached", false))
			events, err = query.List(c.Request().Context(), h.db, c.QueryParam("search"))
		} else {
			didFetch := false
			events, err = h.eventsCache.Fetch(c.Request().URL.String(), func() ([]*domain.Event, error) {
				didFetch = true
				return query.List(c.Request().Context(), h.db, c.QueryParam("search"))
			})
			slogecho.AddCustomAttributes(c, slog.Bool("cached", !didFetch))
		}

		if err != nil {
			if !errors.Is(err, wreck.NotFound) {
				return err
			}
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(offset, events, c.Get("csrf").(string), c.Path()).Render(c.Response())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	s := loadSettings(c)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    s.Title,
		SectionTitle: sectionTitle,
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

// Register the handler.
func (h *EventsHandler) Register(g *echo.Group) {
	g.GET("/", h.Latest)
	g.POST("/", h.Latest) // Fox htmx.
	g.GET("/tag/:tagName", h.Latest)
	g.POST("/tag/:tagName", h.Latest) // For htmx.

	g.GET("/upcoming", h.Upcoming)
	g.POST("/upcoming", h.Upcoming) // Fox htmx.
	g.GET("/upcoming/tag/:tagName", h.Upcoming)
	g.POST("/upcoming/tag/:tagName", h.Upcoming) // For htmx.

	g.GET("/past", h.Past)
	g.POST("/past", h.Past) // For htmx.
	g.GET("/past/tag/:tagName", h.Past)
	g.POST("/past/tag/:tagName", h.Past) // For htmx.

	g.GET("/tags", h.Tags)
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(
	db *bun.DB,
	tagsCache *evcache.Cache[string, []*domain.Tag],
	eventsCache *evcache.Cache[string, []*domain.Event],
) *EventsHandler {
	return &EventsHandler{
		db:          db,
		tagsCache:   tagsCache,
		eventsCache: eventsCache,
	}
}

func getTagFilter(c echo.Context) string {
	v, _ := url.QueryUnescape(c.Param("tagName"))
	return v
}

func getIntQuery(key string, c echo.Context) int64 {
	v, _ := strconv.ParseInt(c.QueryParam(key), 10, 64)
	return v
}
