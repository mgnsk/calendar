package handler

import (
	"net/http"
	"net/url"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/uptrace/bun"
)

// Register handlers.
func Register(
	e *echo.Echo,
	db *bun.DB,
	sm *scs.SessionManager,
	baseURL *url.URL,
) {
	g := e.Group("",
		echo.WrapMiddleware(sm.LoadAndSave),
		LoadSettingsMiddleware(db),
		LoadUserMiddleware(db, sm),
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

	// Setup.
	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		h := NewSetupHandler(db)
		h.Register(g)
	}

	// Authentication.
	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		h := NewAuthenticationHandler(db, sm)
		h.Register(g)
	}

	// Events.
	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		h := NewEventsHandler(db)
		h.Register(g)
	}

	// Events management.
	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		h := NewAddEventHandler(db)
		h.Register(g)
	}

	// Feeds.
	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		h := NewFeedHandler(db, baseURL)
		h.Register(g)
	}
}
