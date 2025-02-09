package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	slogecho "github.com/samber/slog-echo"
	"github.com/uptrace/bun"
)

// AssetCacheMiddleware enables caching for responses.
func AssetCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "max-age=31536000, immutable")

		return next(c)
	}
}

// LoadSettingsMiddleware loads settings or redirects to setup page.
func LoadSettingsMiddleware(db *bun.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			settings, err := model.GetSettings(c.Request().Context(), db)
			if err != nil {
				if !errors.Is(err, wreck.NotFound) {
					return err
				}
			}

			if settings != nil {
				c.Set("settings", settings)
			}

			if c.Path() == "/setup" {
				return next(c)
			}

			if settings != nil {
				return next(c)
			}

			return c.Redirect(http.StatusSeeOther, "/setup")
		}
	}
}

// LoadUserMiddleware loads the current user.
func LoadUserMiddleware(db *bun.DB, sm *scs.SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			username := sm.GetString(c.Request().Context(), "username")
			if username != "" {
				user, err := model.GetUser(c.Request().Context(), db, username)
				if err != nil {
					if !errors.Is(err, wreck.NotFound) {
						return err
					}
				} else {
					c.Set("user", user)
				}
			}

			return next(c)
		}
	}
}

// ErrorHandler handles setting response status code from error and renders an error page.
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			return HandleError(err, c)
		}
	}
}

// HandleError is a custom function to handle errors.
func HandleError(err error, c echo.Context) error {
	var (
		msg  string
		code = http.StatusInternalServerError
	)

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		code = http.StatusRequestTimeout
	} else if werr := *new(wreck.Error); errors.As(err, &werr) {
		msg = werr.Message()
		if v := wreck.Value(werr, wreck.KeyHTTPCode); v != nil {
			code = v.(int)
		}
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	reqID := c.Response().Header().Get(echo.HeaderXRequestID)

	slogecho.AddCustomAttributes(c, slog.String("error", err.Error()))
	c.Response().Status = code

	return html.ErrorPage("Error", code, msg, reqID).Render(c.Response())
}
