package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/ggicci/httpin"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/mgnsk/webkit"
	"github.com/mgnsk/webkit/session"
	"github.com/mgnsk/webkit/web"
	"github.com/uptrace/bun"
)

// AuthenticationHandler handles user login and logout.
type AuthenticationHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Login handles login page.
func (h *AuthenticationHandler) Login(w http.ResponseWriter, r *http.Request) {
	user := server.UserKey.Get(r.Context())
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	switch r.Method {
	case http.MethodGet:
		server.RenderPage(w, r, h.sm,
			html.LoginMain(contract.LoginForm{}, nil),
		)

	case http.MethodPost:
		req := contract.LoginForm{}
		if err := httpin.DecodeTo(r, &req); err != nil {
			panic(err)
		}

		if errs := req.Validate(); len(errs) > 0 {
			// TODO: redirect with flash errors?
			server.RenderPage(w, r, h.sm,
				html.LoginMain(contract.LoginForm{Username: req.Username}, errs),
			)
			return
		}

		// Grace timeout for login failures so we always fail in constant time
		// regardless of whether user does not exist or invalid password provided.
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		user, err := model.GetUserByUsername(ctx, h.db, req.Username)
		if err != nil {
			if errors.Is(err, webkit.NotFound) {
				<-ctx.Done()

				errs := url.Values{}
				errs.Set("username", "Invalid username or password")
				errs.Set("password", "Invalid username or password")

				server.RenderPage(w, r, h.sm,
					html.LoginMain(contract.LoginForm{Username: req.Username}, errs),
				)
				return
			}
			panic(err)
		}

		if err := user.VerifyPassword(req.Password); err != nil {
			if errors.Is(err, webkit.InvalidValue) {
				<-ctx.Done()

				errs := url.Values{}
				errs.Set("username", "Invalid username or password")
				errs.Set("password", "Invalid username or password")

				server.RenderPage(w, r, h.sm,
					html.LoginMain(contract.LoginForm{Username: req.Username}, errs),
				)
				return
			}
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
		panic(webkit.NotFound.New("Not found"))
	}
}

// Logout handles logout page.
func (h *AuthenticationHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sm.Destroy(r.Context()); err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Register the handler.
func (h *AuthenticationHandler) Register(mux *http.ServeMux) {
	middlewares := []web.Middleware{
		session.Middleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /login", web.Chain(http.HandlerFunc(h.Login), middlewares...))
	mux.Handle("POST /login", web.Chain(http.HandlerFunc(h.Login), middlewares...))

	mux.Handle("GET /logout", web.Chain(http.HandlerFunc(h.Logout), middlewares...))
}

// NewAuthenticationHandler creates a new authentication handler.
func NewAuthenticationHandler(db *bun.DB, sm *scs.SessionManager) *AuthenticationHandler {
	return &AuthenticationHandler{
		db: db,
		sm: sm,
	}
}
