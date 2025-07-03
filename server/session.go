package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
)

// NewSessionManager creates a new session manager.
func NewSessionManager(store scs.Store) *scs.SessionManager {
	sm := scs.New()
	sm.Store = store
	sm.HashTokenInStore = true
	sm.Lifetime = 12 * 30 * 24 * time.Hour // 12 months
	sm.IdleTimeout = 30 * 24 * time.Hour   // 30 days
	sm.Cookie.Name = "session_id"
	sm.Cookie.Domain = ""
	sm.Cookie.HttpOnly = true
	sm.Cookie.Path = "/"
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode
	sm.Cookie.Secure = true

	sm.ErrorFunc = func(_ http.ResponseWriter, _ *http.Request, err error) {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}

		// This panic is caught by our custom middleware wrapper
		// and goes through Echo's error handling.
		panic(err)
	}

	return sm
}

// NewSessionMiddleware creates a new session middleware.
func NewSessionMiddleware(sm *scs.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m := wrapRecovery(c.Echo(), sm.LoadAndSave)
			return echo.WrapMiddleware(m)(next)(c)
		}
	}
}

// middlewareFunc is a go std middleware function.
type middlewareFunc func(http.Handler) http.Handler

// wrapRecovery wraps a go std middleware with panic recovery
// and error handling from Echo.
func wrapRecovery(e *echo.Echo, mw middlewareFunc) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := mw(next)

			defer func() {
				if err := recover(); err != nil {
					c := e.NewContext(r, w)
					e.Router().Find(r.Method, echo.GetPath(r), c)
					switch err := err.(type) {
					case error:
						c.Error(err)
					default:
						c.Error(fmt.Errorf("%v", err))
					}
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}
