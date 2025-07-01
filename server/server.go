package server

import (
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/handler"
	"github.com/uptrace/bun"
)

// NewServer creates a new calendar server.
func NewServer(db *bun.DB, sm *scs.SessionManager, baseURL *url.URL) *echo.Echo {
	e := echo.New()

	e.Use(
		handler.Recover(),

		middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; connect-src 'self' *.openstreetmap.org",
			HSTSPreloadEnabled:    false,
		}),

		middleware.RequestID(),

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

	// Static assets.
	calendar.RegisterAssetsHandler(e)

	// Setup.
	{
		g := e.Group("",
			echo.WrapMiddleware(sm.LoadAndSave),
		)

		h := handler.NewSetupHandler(db, sm)
		h.Register(g)
	}

	// Authentication.
	{
		g := e.Group("",
			echo.WrapMiddleware(sm.LoadAndSave),
		)

		h := handler.NewAuthenticationHandler(db, sm)
		h.Register(g)
	}

	// Events.
	{
		g := e.Group("",
			echo.WrapMiddleware(sm.LoadAndSave),
		)

		h := handler.NewEventsHandler(db, sm)
		h.Register(g)
	}

	// Events management.
	{
		g := e.Group("",
			echo.WrapMiddleware(sm.LoadAndSave),
		)

		h := handler.NewEditEventHandler(db, sm)
		h.Register(g)
	}

	// Users management.
	{
		g := e.Group("",
			echo.WrapMiddleware(sm.LoadAndSave),
		)

		h := handler.NewUsersHandler(db, sm)
		h.Register(g)
	}

	// Feeds.
	{
		g := e.Group("",
			echo.WrapMiddleware(handler.NoCache),
		)

		h := handler.NewFeedHandler(db, baseURL)
		h.Register(g)
	}

	e.Server.ReadHeaderTimeout = time.Minute
	e.Server.ReadTimeout = time.Minute
	e.Server.WriteTimeout = time.Minute
	e.Server.IdleTimeout = time.Minute

	return e
}
