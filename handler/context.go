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
	c echo.Context

	Session  *scs.SessionManager
	User     *domain.User
	Settings *domain.Settings
	CSRF     string
}

// Request returns the current request.
func (c *Context) Request() *http.Request {
	return c.c.Request()
}

// Response returns the current response.
func (c *Context) Response() http.ResponseWriter {
	return c.c.Response().Writer
}

// Redirect redirects the request to a provided URL with status code.
func (c *Context) Redirect(code int, url string) error {
	return c.c.Redirect(code, url)
}

// Path returns the registered path for the handler.
func (c *Context) Path() string {
	return c.c.Path()
}

// Bind binds path params, query params and the request body into provided type `dst`. The default binder
// binds body based on Content-Type header.
func (c *Context) Bind(dst any) error {
	return c.c.Bind(dst)
}

// Func defines a function to serve HTTP requests, using the custom context.
type Func func(*Context) error

// Wrap a HandlerFunc with echo.HandlerFunc.
func Wrap(db *bun.DB, sm *scs.SessionManager, next Func) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := &Context{
			c:       c,
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
