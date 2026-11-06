package server

import (
	"context"

	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/webkit/web"
)

// UserKey is the request context key of the signed-in user.
var UserKey web.ContextKey[domain.User]

var settingsKey web.ContextKey[domain.Settings]

// GetSettings returns the settings from context. It returns nil when the
// settings middleware found no settings (before first setup).
func GetSettings(ctx context.Context) *domain.Settings {
	if settings, ok := settingsKey.Lookup(ctx); ok {
		return settings
	}
	// TODO
	return domain.NewDefaultSettings()
}
