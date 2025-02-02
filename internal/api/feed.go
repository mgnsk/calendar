package api

import (
	"encoding/xml"
	"net/http"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/uptrace/bun"
)

// TODO: consider adding an ETag header so that proxies can cache the response.

// FeedHandler handles feed output.
type FeedHandler struct {
	db     *bun.DB
	config Config
}

// Register the handler.
func (h *FeedHandler) Register(e *echo.Echo) {
	e.GET("/feed", h.HandleRSS)
	e.GET("/ical", h.HandleICal)
}

// HandleRSS handles RSS feeds.
func (h *FeedHandler) HandleRSS(c echo.Context) error {
	return h.handleRSSFeed(c, "rss")
}

// HandleICal handles iCal feeds.
func (h *FeedHandler) HandleICal(c echo.Context) error {
	// Upcoming events in created at ASC.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now(), time.Time{}, "", model.OrderCreatedAtAsc, 0, 0)
	if err != nil {
		return err
	}

	cal := ics.NewCalendar()
	cal.SetProductId("Calendar - github.com/mgnsk/calendar")
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription(h.config.PageTitle)
	cal.SetUrl(h.config.BaseURL.JoinPath("/ical").String())

	for _, ev := range events {
		event := cal.AddEvent(ev.ID.String())

		event.SetCreatedTime(ev.GetCreatedAt())
		event.SetModifiedAt(ev.GetCreatedAt())

		event.SetStartAt(ev.StartAt.Time())
		if endAt := ev.EndAt.Time(); !endAt.IsZero() {
			event.SetEndAt(endAt)
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
	// Latest upcoming events in created at ASC.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now(), time.Time{}, "", model.OrderCreatedAtAsc, 0, 0)
	if err != nil {
		return err
	}

	feed := &feeds.Feed{
		Title: h.config.PageTitle,
		Link:  &feeds.Link{Rel: "self", Href: h.config.BaseURL.JoinPath(c.Path()).String()},
		Image: nil,
	}

	for _, ev := range events {
		feed.Add(&feeds.Item{
			Title:       ev.Title,
			Link:        &feeds.Link{Href: ev.URL},
			Description: ev.GetDescription(),
			Id:          ev.ID.String(),
			IsPermaLink: "false",
			Updated:     ev.StartAt.Time(),
			Created:     ev.StartAt.Time(),
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

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB, config Config) *FeedHandler {
	return &FeedHandler{
		db:     db,
		config: config,
	}
}
