package server

import (
	"context"

	"github.com/mgnsk/calendar/domain"
)

type userCtxKey struct{}

// GetUser returns the current user from context if present.
func GetUser(ctx context.Context) *domain.User {
	if user := ctx.Value(userCtxKey{}); user != nil {
		return user.(*domain.User)
	}
	return nil
}

type settingsCtxKey struct{}

// GetSettings returns the settings from context.
func GetSettings(ctx context.Context) *domain.Settings {
	if settings := ctx.Value(settingsCtxKey{}); settings != nil {
		return settings.(*domain.Settings)
	}
	// TODO
	return domain.NewDefaultSettings()
}
