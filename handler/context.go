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

// Context is the request context.
type Context struct {
	echo.Context

	Session  *scs.SessionManager
	User     *domain.User
	Settings *domain.Settings
	CSRF     string
}

// Func defines a function to serve HTTP requests, using the custom context.
type Func func(*Context) error

// Wrap a HandlerFunc with echo.HandlerFunc.
func Wrap(db *bun.DB, sm *scs.SessionManager, next Func) echo.HandlerFunc {
	return func(c echo.Context) error {
		var csrf string
		if v, ok := c.Get("csrf").(string); ok {
			csrf = v
		}

		ctx := &Context{
			Context: c,
			Session: sm,
			CSRF:    csrf,
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

		if sm == nil {
			return next(ctx)
		}

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

		return next(ctx)
	}
}
