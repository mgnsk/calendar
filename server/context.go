package server

import (
	"context"

	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/webkit/web"
)

var (
	userKey     web.ContextKey[domain.User]
	settingsKey web.ContextKey[domain.Settings]
)

// GetUser returns the current user from context if present.
func GetUser(ctx context.Context) *domain.User {
	return userKey.Get(ctx)
}

// GetSettings returns the settings from context. It returns nil when the
// settings middleware found no settings (before first setup).
func GetSettings(ctx context.Context) *domain.Settings {
	if settings, ok := settingsKey.Lookup(ctx); ok {
		return settings
	}
	// TODO
	return domain.NewDefaultSettings()
}
