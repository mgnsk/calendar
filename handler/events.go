package handler

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/ggicci/httpin"
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
func (h *EventsHandler) Upcoming(w http.ResponseWriter, r *http.Request) {
	h.events(
		w,
		r,
		model.NewEventsQuery().WithStartAtFrom(time.Now()),
		model.OrderStartAtAsc,
	)
}

// Past handles past events.
func (h *EventsHandler) Past(w http.ResponseWriter, r *http.Request) {
	h.events(
		w,
		r,
		model.NewEventsQuery().WithStartAtUntil(time.Now()),
		model.OrderStartAtDesc,
	)
}

// MyEvents handles current user events.
func (h *EventsHandler) MyEvents(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	h.events(
		w,
		r,
		model.NewEventsQuery().WithUserID(user.ID).WithIncludeDrafts(),
		model.OrderCreatedAtDesc,
	)
}

// Tags handles tags.
func (h *EventsHandler) Tags(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && hxhttp.IsRequest(r.Header) {
		tags, err := model.ListTags(r.Context(), h.db, time.Now(), 500)
		if err != nil {
			if !errors.Is(err, calendar.NotFound) {
				panic(err)
			}
		}

		slices.SortFunc(tags, func(a, b *domain.Tag) int {
			return strings.Compare(a.Name, b.Name)
		})

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := html.TagListPartial(tags).Render(w); err != nil {
			panic(err)
		}
		return
	}

	server.RenderPage(w, r, h.sm,
		html.TagsMain(),
	)
}

func (h *EventsHandler) events(w http.ResponseWriter, r *http.Request, query model.EventsQueryBuilder, order model.EventOrder) {
	user := server.GetUser(r.Context())

	if r.Method == http.MethodPost && hxhttp.IsRequest(r.Header) {
		req := contract.ListEventsRequest{}
		if err := httpin.DecodeTo(r, &req); err != nil {
			panic(err)
		}

		var cursor int64

		switch r.URL.Path {
		case "/my-events":
			cursor = req.LastID

		case "/", "/past":
			cursor = req.Offset

		default:
			panic(calendar.NotFound.New("Not found"))
		}

		query = query.
			WithOrder(cursor, order).
			WithLimit(contract.EventLimitPerPage).
			WithSearchText(req.Search)

		var (
			events []*domain.Event
			err    error
		)

		events, err = query.List(r.Context(), h.db)
		if err != nil {
			if !errors.Is(err, calendar.NotFound) {
				panic(err)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := html.EventListPartial(user, cursor, events).Render(w); err != nil {
			panic(err)
		}
		return
	}

	server.RenderPage(w, r, h.sm,
		html.EventsMain(),
	)
}

// Register the handler.
func (h *EventsHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /{$}", server.WithMiddleware(http.HandlerFunc(h.Upcoming), middlewares...))
	mux.Handle("POST /{$}", server.WithMiddleware(http.HandlerFunc(h.Upcoming), middlewares...)) // For htmx.

	mux.Handle("GET /past", server.WithMiddleware(http.HandlerFunc(h.Past), middlewares...))
	mux.Handle("POST /past", server.WithMiddleware(http.HandlerFunc(h.Past), middlewares...)) // For htmx.

	mux.Handle("GET /tags", server.WithMiddleware(http.HandlerFunc(h.Tags), middlewares...))
	mux.Handle("POST /tags", server.WithMiddleware(http.HandlerFunc(h.Tags), middlewares...)) // For htmx.

	mux.Handle("GET /my-events", server.WithMiddleware(http.HandlerFunc(h.MyEvents), middlewares...))
	mux.Handle("POST /my-events", server.WithMiddleware(http.HandlerFunc(h.MyEvents), middlewares...)) // For htmx.
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(db *bun.DB, sm *scs.SessionManager) *EventsHandler {
	return &EventsHandler{
		db: db,
		sm: sm,
	}
}
