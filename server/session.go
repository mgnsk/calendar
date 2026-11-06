package server

import (
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/mgnsk/webkit/session"
)

// NewSessionManager creates a new session manager.
func NewSessionManager(store scs.Store) *scs.SessionManager {
	return session.New(store, session.Options{
		Secure:      true,
		Lifetime:    12 * 30 * 24 * time.Hour, // 12 months
		IdleTimeout: 30 * 24 * time.Hour,      // 30 days
	})
}
