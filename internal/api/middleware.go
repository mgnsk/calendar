package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// AssetCacheMiddleware enables caching for responses.
func AssetCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "max-age=31536000, immutable")

		return next(c)
	}
}

// TimeoutMiddleware enables timeout for request contexts.
func TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, cancel := context.WithTimeout(c.Request().Context(), timeout)
			defer cancel()

			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
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

			return c.Redirect(http.StatusFound, "/setup")
		}
	}
}
