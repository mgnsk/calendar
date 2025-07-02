package server

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
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

	return sm
}
