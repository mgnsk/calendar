package api

import (
	"fmt"
	"net/http"

	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/uptrace/bun"
)

// TODO: consider adding an ETag header so that proxies can cache the response.

// Paths for endpoints.
const (
	RSSPath  = "/feed"
	AtomPath = "/feed/atom"
	ICalPath = "/ical"
)

// FeedHandler handles feed output.
type FeedHandler struct {
	db     *bun.DB
	config FeedConfig
}

// Register the handler.
func (h *FeedHandler) Register(e *echo.Echo) {
	e.GET(RSSPath, h.HandleRSS)
	e.GET(AtomPath, h.HandleAtom)
	e.GET(ICalPath, h.HandleICal)
}

// HandleRSS handles RSS feeds.
func (h *FeedHandler) HandleRSS(c echo.Context) error {
	return h.handleRSSFeed(c, "rss")
}

// HandleAtom handles Atom feeds.
func (h *FeedHandler) HandleAtom(c echo.Context) error {
	return h.handleRSSFeed(c, "atom")
}

// HandleICal handles iCal feeds.
func (h *FeedHandler) HandleICal(c echo.Context) error {
	events, err := model.ListEvents(c.Request().Context(), h.db, "asc")
	if err != nil {
		return err
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetDescription(h.config.Title)
	cal.SetUrl(h.config.BaseURL.JoinPath(ICalPath).String())

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
	c.Response().WriteHeader(http.StatusOK)

	return cal.SerializeTo(c.Response())
}

func (h *FeedHandler) handleRSSFeed(c echo.Context, target string) error {
	events, err := model.ListEvents(c.Request().Context(), h.db, "asc")
	if err != nil {
		return err
	}

	feed := &feeds.Feed{
		Title: h.config.Title,
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

	c.Response().Header().Set(echo.HeaderContentDisposition, `attachment; filename="feed.xml"`)

	switch target {
	case "rss":
		c.Response().Header().Set(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
		c.Response().WriteHeader(http.StatusOK)
		return feed.WriteRss(c.Response())

	case "atom":
		c.Response().Header().Set(echo.HeaderContentType, "application/atom+xml; charset=utf-8")
		c.Response().WriteHeader(http.StatusOK)
		return feed.WriteAtom(c.Response())

	default:
		panic(fmt.Sprintf("invalid feed type %v", target))
	}
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB, config FeedConfig) *FeedHandler {
	return &FeedHandler{
		db:     db,
		config: config,
	}
}
