package handler

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
)

// FeedHandler handles feed output.
type FeedHandler struct {
	db      *bun.DB
	baseURL *url.URL
}

// HandleRSS handles RSS feeds.
func (h *FeedHandler) HandleRSS(c *server.Context) error {
	return h.handleRSSFeed(c, "rss")
}

// HandleICal handles iCal feeds.
func (h *FeedHandler) HandleICal(c *server.Context) error {
	events, err := h.getEvents(c)
	if err != nil {
		return err
	}

	cal := ics.NewCalendar()
	cal.SetProductId("Calendar - github.com/mgnsk/calendar")
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription(c.Settings.Title)
	cal.SetUrl(h.baseURL.JoinPath("/calendar.ics").String())

	for _, ev := range events {
		event := cal.AddEvent(ev.ID.String())

		event.SetCreatedTime(ev.GetCreatedAt())
		event.SetModifiedAt(ev.GetCreatedAt())

		event.SetStartAt(ev.StartAt)
		// Default to 1 hour event duration.
		event.SetEndAt(ev.StartAt.Add(time.Hour))

		event.SetSummary(ev.Title)
		event.SetURL(ev.URL)
		event.SetDescription(ev.Description)
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/calendar; charset=utf-8")
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="calendar.ics"`)
	c.Response().Header().Set(echo.HeaderCacheControl, "max-age=3600")
	c.Response().Header().Set("Expires", time.Now().UTC().Add(time.Hour).Format(http.TimeFormat))

	c.Response().WriteHeader(http.StatusOK)

	return cal.SerializeTo(c.Response())
}

func (h *FeedHandler) handleRSSFeed(c *server.Context, _ string) error {
	events, err := h.getEvents(c)
	if err != nil {
		return err
	}

	feed := &feeds.Feed{
		Title: c.Settings.Title,
		Link:  &feeds.Link{Rel: "self", Href: h.baseURL.JoinPath(c.Path()).String()},
		Image: nil,
	}

	for _, ev := range events {
		var htmlContent strings.Builder
		if err := html.EventCard(nil, ev, "").Render(&htmlContent); err != nil {
			return err
		}

		feed.Add(&feeds.Item{
			Title:       ev.Title,
			Link:        &feeds.Link{Href: ev.URL},
			Description: fmt.Sprintf("%s\n\n%s", ev.GetDateString(), ev.Description),
			Content:     htmlContent.String(),
			Id:          ev.ID.String(),
			IsPermaLink: "false",
			Updated:     ev.GetCreatedAt(),
			Created:     ev.GetCreatedAt(),
		})
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="feed.rss"`)
	c.Response().Header().Set(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
	c.Response().Header().Set(echo.HeaderCacheControl, "max-age=3600")
	c.Response().Header().Set("Expires", time.Now().UTC().Add(time.Hour).Format(http.TimeFormat))
	c.Response().WriteHeader(http.StatusOK)

	rss := (&feeds.Rss{Feed: feed}).RssFeed()
	rss.Generator = "Calendar - github.com/mgnsk/calendar"
	x := rss.FeedXml()

	// write default xml header, without the newline
	if _, err := c.Response().Write([]byte(xml.Header[:len(xml.Header)-1])); err != nil {
		return err
	}

	e := xml.NewEncoder(c.Response())
	e.Indent("", "  ")

	return e.Encode(x)
}

func (h *FeedHandler) getEvents(c *server.Context) ([]*domain.Event, error) {
	return model.NewEventsQuery().
		WithStartAtFrom(time.Now()).
		WithOrder(0, model.OrderCreatedAtAsc).
		WithLimit(100). // TODO: test this with rss and thunderbird calendar.
		List(c.Request().Context(), h.db, "")
}

// Register the handler.
func (h *FeedHandler) Register(g *echo.Group) {
	g.GET("/feed", server.Wrap(h.db, nil, h.HandleRSS))
	g.GET("/calendar.ics", server.Wrap(h.db, nil, h.HandleICal))
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB, baseURL *url.URL) *FeedHandler {
	return &FeedHandler{
		db:      db,
		baseURL: baseURL,
	}
}
