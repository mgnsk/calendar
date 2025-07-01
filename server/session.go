package server

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
)

// NewSessionManager creates a new session manager.
func NewSessionManager(store scs.Store, e *echo.Echo) *scs.SessionManager {
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

	sm.ErrorFunc = func(w http.ResponseWriter, r *http.Request, err error) {
		c := e.NewContext(r, w)
		c.Error(err)
	}

	return sm
}
