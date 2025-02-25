package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/wreck"
	slogecho "github.com/samber/slog-echo"
	"github.com/uptrace/bun"
)

// GetContext extracts the custom context.
func GetContext(c echo.Context) Context {
	return c.Get("context").(Context)
}

// Context is the request context.
type Context struct {
	Session  *scs.SessionManager
	User     *domain.User
	Settings *domain.Settings
	CSRF     string
}

// SetContextMiddleware sets custom request context.
// Must be registered after session and CSRF middleware.
func SetContextMiddleware(db *bun.DB, sm *scs.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := Context{
				Session: sm,
				CSRF:    c.Get("csrf").(string),
			}

			settings, err := model.GetSettings(c.Request().Context(), db)
			if err != nil {
				if !errors.Is(err, wreck.NotFound) {
					return err
				}
			}

			if settings == nil && c.Path() != "/setup" {
				return c.Redirect(http.StatusSeeOther, "/setup")
			}

			ctx.Settings = settings

			if username := sm.GetString(c.Request().Context(), "username"); username != "" {
				user, err := model.GetUser(c.Request().Context(), db, username)
				if err != nil {
					if !errors.Is(err, wreck.NotFound) {
						return err
					}
				}

				if user == nil {
					if err := sm.Destroy(c.Request().Context()); err != nil {
						return err
					}
					return c.Redirect(http.StatusSeeOther, "/")
				}

				ctx.User = user
				slogecho.AddCustomAttributes(c, slog.String("username", username))
			}

			c.Set("context", ctx)

			return next(c)
		}
	}
}
