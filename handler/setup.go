package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

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
)

// SetupHandler handles setup pages.
type SetupHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Setup handles the setup page.
func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	settings := server.GetSettings(r.Context())
	if settings != nil {
		// Already set up.
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	settings = domain.NewDefaultSettings()

	switch r.Method {
	case http.MethodGet:
		form := contract.SetupForm{
			Title:       settings.Title,
			Description: settings.Description,
		}

		server.RenderPage(w, r, h.sm,
			html.SetupMain(form, nil),
		)
		return

	case http.MethodPost:
		req := contract.SetupForm{}
		if err := httpin.DecodeTo(r, &req); err != nil {
			panic(err)
		}

		if errs := req.Validate(); len(errs) > 0 {
			server.RenderPage(w, r, h.sm,
				html.SetupMain(req, errs),
			)
			return
		}

		settings.Title = req.Title
		settings.Description = req.Description

		user := &domain.User{
			ID:       snowflake.Generate(),
			Username: req.Username,
			Role:     domain.Admin,
		}

		if err := user.SetPassword(req.Password1); err != nil {
			if errors.Is(err, calendar.InvalidValue) {
				errs := url.Values{}
				errs.Set("password1", err.Error())
				errs.Set("password2", err.Error())

				server.RenderPage(w, r, h.sm,
					html.SetupMain(req, errs),
				)
				return
			}

			panic(err)
		}

		if err := h.db.RunInTx(r.Context(), nil, func(ctx context.Context, db bun.Tx) error {
			if err := model.InsertSettings(ctx, db, settings); err != nil {
				return err
			}

			return model.InsertUser(ctx, db, user)
		}); err != nil {
			panic(err)
		}

		// First renew the session token.
		if err := h.sm.RenewToken(r.Context()); err != nil {
			panic(err)
		}

		// Then make the privilege-level change.
		h.sm.Put(r.Context(), "username", user.Username)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	default:
		panic(calendar.NotFound.New("Not found"))
	}
}

// Register the handler.
func (h *SetupHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
	}

	mux.Handle("GET /setup", server.WithMiddleware(http.HandlerFunc(h.Setup), middlewares...))
	mux.Handle("POST /setup", server.WithMiddleware(http.HandlerFunc(h.Setup), middlewares...))
}

// NewSetupHandler creates a new setup handler.
func NewSetupHandler(db *bun.DB, sm *scs.SessionManager) *SetupHandler {
	return &SetupHandler{
		db: db,
		sm: sm,
	}
}
