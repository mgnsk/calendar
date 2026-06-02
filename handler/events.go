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
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// EventsHandler handles event pages rendering.
type EventsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Events handles event list page.
func (h *EventsHandler) Events(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())

	req := contract.ListEventsRequest{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	var (
		startAtFrom, startAtUntil time.Time
		filterUserID              snowflake.ID
		cursor                    int64
		order                     model.EventOrder
	)

	var query model.EventsQueryBuilder

	switch r.URL.Path {
	case "/":
		startAtFrom = time.Now()
		cursor = req.Offset
		order = model.OrderStartAtAsc
		query = model.NewEventsQuery().WithStartAtFrom(startAtFrom)

	case "/past":
		startAtUntil = time.Now()
		cursor = req.Offset
		order = model.OrderStartAtDesc
		query = model.NewEventsQuery().WithStartAtUntil(startAtUntil)

	case "/my-events":
		if user == nil {
			panic(calendar.Forbidden.New("Must be logged in"))
		}

		filterUserID = user.ID
		cursor = req.LastID
		order = model.OrderCreatedAtDesc
		query = model.NewEventsQuery().WithUserID(filterUserID).WithIncludeDrafts()

	default:
		panic(calendar.Internal.New("Unhandled path"))
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

	if hxhttp.IsRequest(r.Header) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if err := html.EventListPartial(user, cursor, events).Render(w); err != nil {
			panic(err)
		}
		return
	}

	tags, err := model.ListTags(r.Context(), h.db, startAtFrom, startAtUntil, filterUserID, 500)
	if err != nil {
		if !errors.Is(err, calendar.NotFound) {
			panic(err)
		}
	}

	slices.SortFunc(tags, func(a, b *domain.Tag) int {
		return strings.Compare(a.Name, b.Name)
	})

	server.RenderPage(w, r, h.sm,
		html.EventsMain(user, req.Offset, events, tags),
	)
}

// Register the handler.
func (h *EventsHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /{$}", server.WithMiddleware(http.HandlerFunc(h.Events), middlewares...))
	mux.Handle("GET /past", server.WithMiddleware(http.HandlerFunc(h.Events), middlewares...))
	mux.Handle("GET /my-events", server.WithMiddleware(http.HandlerFunc(h.Events), middlewares...))
}

// NewEventsHandler creates a new events handler.
func NewEventsHandler(db *bun.DB, sm *scs.SessionManager) *EventsHandler {
	return &EventsHandler{
		db: db,
		sm: sm,
	}
}
