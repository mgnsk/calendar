package calendar

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
)

// RegisterAssetsHandler registers the static assets echo handler.
func RegisterAssetsHandler(e *echo.Echo) {
	e.GET("/assets/*",
		echo.StaticDirectoryHandler(assetsFS, false),
		assetCacheMiddleware(365*24*time.Hour),
	)
}

func assetCacheMiddleware(d time.Duration) echo.MiddlewareFunc {
	value := fmt.Sprintf("max-age=%d, immutable", int64(d.Seconds()))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", value)

			return next(c)
		}
	}
}
