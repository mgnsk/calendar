package server

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// NewServer creates a new calendar server.
func NewServer() *echo.Echo {
	e := echo.New()

	e.HTTPErrorHandler = ErrorHandler()

	e.Use(
		ErrorLogger(),
		Recover(),

		middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
			ErrorHandler: func(err error, _ echo.Context) error {
				// Pass through error.
				return err
			},
			Timeout: time.Second,
		}),

		middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; connect-src 'self' *.openstreetmap.org",
			HSTSPreloadEnabled:    false,
		}),

		middleware.BodyLimit("1M"),

		middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLength:    32,
			TokenLookup:    "form:csrf",
			ContextKey:     "csrf",
			CookieName:     "_csrf",
			CookieDomain:   "",
			CookiePath:     "/",
			CookieMaxAge:   86400,
			CookieSecure:   true,
			CookieHTTPOnly: true,
			CookieSameSite: http.SameSiteStrictMode,
		}),
	)

	e.Server.ReadTimeout = time.Minute
	e.Server.WriteTimeout = time.Minute

	return e
}
