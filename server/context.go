package server

import (
	"errors"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/uptrace/bun"
)

// Context is the request context.
type Context struct {
	echo.Context

	User     *domain.User
	Settings *domain.Settings
	CSRF     string
}

// HandlerFunc defines a function to serve HTTP requests, using the custom context.
type HandlerFunc func(*Context) error

// Wrap a HandlerFunc with echo.HandlerFunc.
func Wrap(db *bun.DB, sm *scs.SessionManager, next HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var csrf string
		if v, ok := c.Get("csrf").(string); ok {
			csrf = v
		}

		ctx := &Context{
			Context: c,
			CSRF:    csrf,
		}

		settings, err := model.GetSettings(c.Request().Context(), db)
		if err != nil {
			if !errors.Is(err, calendar.NotFound) {
				return err
			}
		}

		if settings == nil && c.Path() != "/setup" {
			// First setup.
			return c.Redirect(http.StatusSeeOther, "/setup")
		}

		ctx.Settings = settings

		if sm == nil {
			// Public endpoint.
			return next(ctx)
		}

		if username := sm.GetString(c.Request().Context(), "username"); username != "" {
			user, err := model.GetUserByUsername(c.Request().Context(), db, username)
			if err != nil {
				if !errors.Is(err, calendar.NotFound) {
					return err
				}
			}

			if user == nil {
				// User has been deleted.
				if err := sm.Destroy(c.Request().Context()); err != nil {
					return err
				}
				return c.Redirect(http.StatusSeeOther, "/")
			}

			ctx.User = user
		}

		return next(ctx)
	}
}
