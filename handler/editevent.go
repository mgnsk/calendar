package handler

import (
	"fmt"
	"net/http"
	"net/url"
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

// TimezoneFinder finds timezone by geo coordinates.
type TimezoneFinder interface {
	GetTimezoneName(lng, lat float64) string
}

// EditEventHandler handles adding and editing events.
type EditEventHandler struct {
	db     *bun.DB
	sm     *scs.SessionManager
	finder TimezoneFinder
}

// Edit handles adding and editing events.
func (h *EditEventHandler) Edit(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	req := contract.EditEventForm{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	var ev *domain.Event

	if req.EventID > 0 {
		event, err := model.GetEvent(r.Context(), h.db, req.EventID)
		if err != nil {
			panic(err)
		}

		if user.Role != domain.Admin && user.ID != event.UserID {
			panic(calendar.Forbidden.New("Non-admin users can only edit own events"))
		}

		ev = event
	}

	switch r.Method {
	case http.MethodGet:
		if ev != nil {
			req.Title = ev.Title
			req.IsDraft = ev.IsDraft
			req.Description = ev.Description
			req.URL = ev.URL
			req.StartAt = ev.StartAt.Format(contract.FormDateTimeLayout)
			req.Location = ev.Location
			req.OSMType = ev.OSMType
			req.OSMID = ev.OSMID
			req.Latitude = ev.Latitude
			req.Longitude = ev.Longitude
		}

		server.RenderPage(w, r, h.sm,
			html.EditEventMain(req, nil),
		)
		return

	case http.MethodPost:
		if errs := req.Validate(); len(errs) > 0 {
			server.RenderPage(w, r, h.sm,
				html.EditEventMain(req, errs),
			)
			return
		}

		startAt, err := h.parseStartAt(req)
		if err != nil {
			errs := url.Values{}
			errs.Set("start_at", "Invalid start_at value")
			server.RenderPage(w, r, h.sm,
				html.EditEventMain(req, errs),
			)
			return
		}

		if ev != nil {
			ev.StartAt = startAt
			ev.Title = req.Title
			ev.IsDraft = req.IsDraft
			ev.Description = req.Description
			ev.URL = req.URL
			ev.Location = req.Location
			ev.OSMType = req.OSMType
			ev.OSMID = req.OSMID
			ev.Latitude = req.Latitude
			ev.Longitude = req.Longitude

			if err := model.UpdateEvent(r.Context(), h.db, ev); err != nil {
				panic(err)
			}

			if req.IsDraft {
				h.sm.Put(r.Context(), "flash-success", "Draft saved")
			} else {
				h.sm.Put(r.Context(), "flash-success", "Event published")
			}

			http.Redirect(w, r, fmt.Sprintf("/edit/%d", ev.ID), http.StatusSeeOther)
			return
		}

		eventID := snowflake.Generate()

		if err := model.InsertEvent(r.Context(), h.db, &domain.Event{
			ID:          eventID,
			StartAt:     startAt,
			Title:       req.Title,
			Description: req.Description,
			URL:         req.URL,
			Location:    req.Location,
			OSMType:     req.OSMType,
			OSMID:       req.OSMID,
			Latitude:    req.Latitude,
			Longitude:   req.Longitude,
			IsDraft:     req.IsDraft,
			UserID:      user.ID,
		}); err != nil {
			panic(err)
		}

		if req.IsDraft {
			h.sm.Put(r.Context(), "flash-success", "Draft saved")
		} else {
			h.sm.Put(r.Context(), "flash-success", "Event published")
		}

		http.Redirect(w, r, fmt.Sprintf("/edit/%d", eventID), http.StatusSeeOther)
		return

	default:
		panic(calendar.NotFound.New("Not found"))
	}
}

// Delete handles deleting events.
func (h *EditEventHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	req := contract.DeleteEventRequest{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	ev, err := model.GetEvent(r.Context(), h.db, req.EventID)
	if err != nil {
		panic(err)
	}

	if user.Role != domain.Admin && user.ID != ev.UserID {
		panic(calendar.Forbidden.New("Non-admin users can only edit own events"))
	}

	if r.Method == http.MethodPost && hxhttp.IsRequest(r.Header) {
		if err := model.DeleteEvent(r.Context(), h.db, ev); err != nil {
			panic(err)
		}

		h.sm.Put(r.Context(), "flash-success", "Event deleted")

		hxhttp.SetRefresh(w.Header())

		return
	}

	panic(calendar.NotFound.New("Not found"))
}

// Preview returns a preview of the event.
func (h *EditEventHandler) Preview(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	req := contract.EditEventForm{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	startAt, _ := h.parseStartAt(req)

	ev := &domain.Event{
		StartAt:     startAt,
		Title:       req.Title,
		Description: req.Description,
		URL:         req.URL,
		Location:    req.Location,
		OSMType:     req.OSMType,
		OSMID:       req.OSMID,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		IsDraft:     req.IsDraft,
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	if err := html.EventCard(nil, ev).Render(w); err != nil {
		panic(err)
	}
}

// Register the handler.
func (h *EditEventHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /edit/{event_id}", server.WithMiddleware(http.HandlerFunc(h.Edit), middlewares...))
	mux.Handle("POST /edit/{event_id}", server.WithMiddleware(http.HandlerFunc(h.Edit), middlewares...))

	mux.Handle("POST /delete/{event_id}", server.WithMiddleware(http.HandlerFunc(h.Delete), middlewares...))

	mux.Handle("POST /preview", server.WithMiddleware(http.HandlerFunc(h.Preview), middlewares...))
}

func (h *EditEventHandler) parseStartAt(req contract.EditEventForm) (time.Time, error) {
	ianaTimezone := h.finder.GetTimezoneName(req.Longitude, req.Latitude)

	if ianaTimezone == "" {
		// If timezone not found, fall back to user timezone.
		ianaTimezone = req.UserTimezone
	}

	var loc *time.Location

	if ianaTimezone == "" {
		// if user timezone also not found, fall back to UTC.
		loc = time.UTC
	} else {
		l, err := time.LoadLocation(ianaTimezone)
		if err != nil {
			// TODO: should we log this error and still use UTC?
			return time.Time{}, calendar.InvalidValue.New("Invalid location timezone", err)
		}

		loc = l
	}

	startAt, err := time.ParseInLocation(contract.FormDateTimeLayout, req.StartAt, loc)
	if err != nil {
		return time.Time{}, calendar.InvalidValue.New("Invalid start_at value", err)
	}

	return startAt, nil
}

// NewEditEventHandler creates a new edit event handler.
func NewEditEventHandler(db *bun.DB, sm *scs.SessionManager, finder TimezoneFinder) *EditEventHandler {
	return &EditEventHandler{
		db:     db,
		sm:     sm,
		finder: finder,
	}
}
