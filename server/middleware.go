package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/webkit/session"
	"github.com/mgnsk/webkit/web"
	"github.com/uptrace/bun"
)

// MiddlewareFunc is a HTTP middleware function.
type MiddlewareFunc = web.Middleware

// WithMiddleware applies middleware to run in order before handler.
var WithMiddleware = web.Chain

// NewSessionMiddleware creates a new session middleware.
func NewSessionMiddleware(sm *scs.SessionManager) MiddlewareFunc {
	return session.Middleware(sm)
}

// NewSettingsMiddleware creates a new settings middleware.
func NewSettingsMiddleware(db *bun.DB) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			settings, err := model.GetSettings(r.Context(), db)
			if err != nil {
				if !errors.Is(err, calendar.NotFound) {
					panic(err)
				}
			}

			if settings == nil && r.URL.Path != "/setup" {
				// First setup.
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
				return
			}

			r = r.WithContext(settingsKey.With(r.Context(), settings))

			next.ServeHTTP(w, r)
		})
	}
}

// NewUserMiddleware creates a new user middleware. A nil sm means a public
// endpoint without a user.
func NewUserMiddleware(db *bun.DB, sm *scs.SessionManager) MiddlewareFunc {
	if sm == nil {
		return func(next http.Handler) http.Handler { return next }
	}

	return session.LoadUser(sm, userKey, func(ctx context.Context, sm *scs.SessionManager) (*domain.User, error) {
		username := sm.GetString(ctx, "username")
		if username == "" {
			return nil, nil
		}
		return model.GetUserByUsername(ctx, db, username)
	}, "/")
}
