package handler

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EventLimitPerPage specifies maximum number of events per page.
const EventLimitPerPage = 3

// EventsHandler handles event pages rendering.
type EventsHandler struct {
	db *bun.DB
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
	if hxhttp.IsRequest(c.Request().Header) {
		tags, err := model.ListTags(c.Request().Context(), h.db)
		if err != nil {
			return err
		}

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.TagListPartial(tags).Render(c.Response())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	user := loadUser(c)
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
	filterTag, err := h.getTagFilter(c)
	if err != nil {
		return err
	}

	if hxhttp.IsRequest(c.Request().Header) {
		lastID, err := h.getIntParam("last_id", c)
		if err != nil {
			return err
		}

		offset, err := h.getIntParam("offset", c)
		if err != nil {
			return err
		}
		if offset > 0 {
			offset += EventLimitPerPage
		}

		cursor := cmp.Or(offset, lastID)

		query = query.
			WithOrder(cursor, order).
			WithFilterTags(filterTag).
			WithLimit(EventLimitPerPage)

		events, err := query.List(c.Request().Context(), h.db, c.FormValue("search"))
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
	user := loadUser(c)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    s.Title,
		SectionTitle: sectionTitle,
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

func (h *EventsHandler) getTagFilter(c echo.Context) (string, error) {
	if param := c.Param("tagName"); param != "" {
		v, err := url.QueryUnescape(param)
		if err != nil {
			return "", wreck.InvalidValue.New("Invalid tag filter", err)
		}
		return v, nil
	}

	return "", nil
}

func (h *EventsHandler) getIntParam(key string, c echo.Context) (int64, error) {
	if v := c.FormValue(key); v != "" {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, wreck.InvalidValue.New(fmt.Sprintf("Invalid %s", key), err)
		}
		return val, nil
	}

	return 0, nil
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
func NewEventsHandler(db *bun.DB) *EventsHandler {
	return &EventsHandler{
		db: db,
	}
}
