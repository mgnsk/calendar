package handler

import (
	"encoding/xml"
	"net/http"
	"net/url"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/uptrace/bun"
)

// FeedHandler handles feed output.
type FeedHandler struct {
	db      *bun.DB
	baseURL *url.URL
}

// HandleRSS handles RSS feeds.
func (h *FeedHandler) HandleRSS(c echo.Context) error {
	return h.handleRSSFeed(c, "rss")
}

// HandleICal handles iCal feeds.
func (h *FeedHandler) HandleICal(c echo.Context) error {
	events, err := h.getEvents(c)
	if err != nil {
		return err
	}

	settings := loadSettings(c)

	cal := ics.NewCalendar()
	cal.SetProductId("Calendar - github.com/mgnsk/calendar")
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription(settings.Title)
	cal.SetUrl(h.baseURL.JoinPath("/ical").String())

	for _, ev := range events {
		event := cal.AddEvent(ev.ID.String())

		event.SetCreatedTime(ev.GetCreatedAt())
		event.SetModifiedAt(ev.GetCreatedAt())

		event.SetStartAt(ev.StartAt)
		if !ev.EndAt.IsZero() {
			event.SetEndAt(ev.EndAt)
		}

		event.SetSummary(ev.Title)
		event.SetURL(ev.URL)
		event.SetDescription(ev.GetDescription())
	}

	c.Response().Header().Set(echo.HeaderContentType, "text/calendar; charset=utf-8")
	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="calendar.ics"`)
	c.Response().Header().Set(echo.HeaderCacheControl, "max-age=3600")
	c.Response().Header().Set("Expires", time.Now().UTC().Add(time.Hour).Format(http.TimeFormat))

	c.Response().WriteHeader(http.StatusOK)

	return cal.SerializeTo(c.Response())
}

func (h *FeedHandler) handleRSSFeed(c echo.Context, _ string) error {
	events, err := h.getEvents(c)
	if err != nil {
		return err
	}

	settings := loadSettings(c)

	feed := &feeds.Feed{
		Title: settings.Title,
		Link:  &feeds.Link{Rel: "self", Href: h.baseURL.JoinPath(c.Path()).String()},
		Image: nil,
	}

	for _, ev := range events {
		feed.Add(&feeds.Item{
			Title:       ev.Title,
			Link:        &feeds.Link{Href: ev.URL},
			Description: ev.GetDescription(),
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

func (h *FeedHandler) getEvents(c echo.Context) ([]*domain.Event, error) {
	return model.NewEventsQuery().
		WithStartAtFrom(time.Now()).
		WithOrder(0, model.OrderCreatedAtAsc).
		WithLimit(100). // TODO: test this with rss and thunderbird calendar.
		List(c.Request().Context(), h.db, "")
}

// Register the handler.
func (h *FeedHandler) Register(g *echo.Group) {
	g.GET("/feed", h.HandleRSS)
	g.GET("/ical", h.HandleICal)
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB, baseURL *url.URL) *FeedHandler {
	return &FeedHandler{
		db:      db,
		baseURL: baseURL,
	}
}
