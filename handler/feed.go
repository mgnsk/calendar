package handler

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/gorilla/feeds"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
)

// FeedHandler handles feed output.
type FeedHandler struct {
	db *bun.DB
}

// HandleRSS handles RSS feeds.
func (h *FeedHandler) HandleRSS(w http.ResponseWriter, r *http.Request) {
	h.handleRSSFeed(w, r, "rss")
}

// HandleICal handles iCal feeds.
func (h *FeedHandler) HandleICal(w http.ResponseWriter, r *http.Request) {
	events, err := h.getEvents(r.Context())
	if err != nil {
		panic(err)
	}

	settings := server.GetSettings(r.Context())

	cal := ics.NewCalendar()
	cal.SetProductId("Calendar - github.com/mgnsk/calendar")
	cal.SetMethod(ics.MethodPublish)
	cal.SetName(settings.Title)
	cal.SetDescription(settings.Description)

	for _, ev := range events {
		event := cal.AddEvent(ev.ID.String())

		event.SetLocation(ev.Location)
		event.SetGeo(ev.Latitude, ev.Longitude)

		event.SetCreatedTime(ev.GetCreatedAt())
		event.SetModifiedAt(ev.GetCreatedAt())
		event.SetDtStampTime(ev.GetCreatedAt())

		event.SetStartAt(ev.StartAt)
		// Default to 1 hour event duration.
		event.SetEndAt(ev.StartAt.Add(time.Hour))

		event.SetSummary(ev.Title)
		event.SetURL(ev.URL)
		event.SetDescription(ev.Description)
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="calendar.ics"`)

	w.WriteHeader(http.StatusOK)

	if err := cal.SerializeTo(w); err != nil {
		panic(err)
	}
}

func (h *FeedHandler) handleRSSFeed(w http.ResponseWriter, r *http.Request, _ string) {
	events, err := h.getEvents(r.Context())
	if err != nil {
		panic(err)
	}

	settings := server.GetSettings(r.Context())

	feed := &feeds.Feed{
		Title:       settings.Title,
		Description: settings.Description,
	}

	for _, ev := range events {
		var htmlContent strings.Builder
		if err := html.EventCard(nil, ev).Render(&htmlContent); err != nil {
			panic(err)
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

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="feed.rss"`)

	w.WriteHeader(http.StatusOK)

	rss := (&feeds.Rss{Feed: feed}).RssFeed()
	rss.Generator = "Calendar - github.com/mgnsk/calendar"
	x := rss.FeedXml()

	// write default xml header, without the newline
	if _, err := w.Write([]byte(xml.Header[:len(xml.Header)-1])); err != nil {
		panic(err)
	}

	e := xml.NewEncoder(w)
	e.Indent("", "  ")

	if err := e.Encode(x); err != nil {
		panic(err)
	}
}

func (h *FeedHandler) getEvents(ctx context.Context) ([]*domain.Event, error) {
	return model.NewEventsQuery().
		WithOrder(0, model.OrderCreatedAtAsc).
		List(ctx, h.db)
}

// Register the handler.
func (h *FeedHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSettingsMiddleware(h.db),
	}

	mux.Handle("GET /feed", server.WithMiddleware(http.HandlerFunc(h.HandleRSS), middlewares...))
	mux.Handle("GET /calendar.ics", server.WithMiddleware(http.HandlerFunc(h.HandleICal), middlewares...))
}

// NewFeedHandler creates a new feed handler.
func NewFeedHandler(db *bun.DB) *FeedHandler {
	return &FeedHandler{
		db: db,
	}
}
