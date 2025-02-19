package calendar

import (
	"github.com/labstack/echo/v4"
)

// RegisterAssetsHandler registers the static assets echo handler.
func RegisterAssetsHandler(e *echo.Echo) {
	e.GET("/assets/*",
		echo.StaticDirectoryHandler(assetsFS, false),
		assetCacheMiddleware,
	)
}

func assetCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "max-age=31536000, immutable")

		return next(c)
	}
}
