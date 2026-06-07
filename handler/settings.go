package handler

import (
	"bufio"
	"context"
	"net/http"
	"strings"

	"github.com/alexedwards/scs/v2"
	"github.com/ggicci/httpin"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
)

// SettingsHandler handles settings pages.
type SettingsHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Settings handles the settings page.
func (h *SettingsHandler) Settings(w http.ResponseWriter, r *http.Request) {
	settings := server.GetSettings(r.Context())
	if settings == nil {
		// TODO: this is already handled by settings middleware
		http.Redirect(w, r, "/setup", http.StatusSeeOther)
		return
	}

	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	if user.Role != domain.Admin {
		panic(calendar.Forbidden.New("Only admins can change settings"))
	}

	switch r.Method {
	case http.MethodGet:
		stopwords, err := model.ListStopWords(r.Context(), h.db)
		if err != nil {
			panic(err)
		}

		form := contract.SettingsForm{
			Title:       settings.Title,
			Description: settings.Description,
			Stopwords:   strings.Join(stopwords, "\n"),
		}

		server.RenderPage(w, r, h.sm,
			html.SettingsMain(form, nil),
		)
		return

	case http.MethodPost:
		req := contract.SettingsForm{}
		if err := httpin.DecodeTo(r, &req); err != nil {
			panic(err)
		}

		if errs := req.Validate(); len(errs) > 0 {
			server.RenderPage(w, r, h.sm,
				html.SettingsMain(req, errs),
			)
			return
		}

		var words []string

		scanner := bufio.NewScanner(strings.NewReader(req.Stopwords))
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			panic(err)
		}

		settings.Title = req.Title
		settings.Description = req.Description

		if err := h.db.RunInTx(r.Context(), nil, func(ctx context.Context, db bun.Tx) error {
			if err := model.UpdateSettings(ctx, db, settings); err != nil {
				return err
			}

			return model.SetStopWords(ctx, db, domain.NewStopWordList(words...))
		}); err != nil {
			panic(err)
		}

		h.sm.Put(r.Context(), "flash-success", "Settings saved")

		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return

	default:
		panic(calendar.NotFound.New("Not found"))
	}
}

// Register the handler.
func (h *SettingsHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /settings", server.WithMiddleware(http.HandlerFunc(h.Settings), middlewares...))
	mux.Handle("POST /settings", server.WithMiddleware(http.HandlerFunc(h.Settings), middlewares...))
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(db *bun.DB, sm *scs.SessionManager) *SettingsHandler {
	return &SettingsHandler{
		db: db,
		sm: sm,
	}
}
