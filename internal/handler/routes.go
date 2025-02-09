package handler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/evcache/v4"
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

	eventsCache := evcache.New[string, []*domain.Event](
		evcache.WithCapacity(128),
		evcache.WithTTL(time.Minute),
		evcache.WithPolicy(evcache.LRU),
	)

	// Events.
	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		tagsCache := evcache.New[string, []*domain.Tag](
			evcache.WithCapacity(128),
			evcache.WithTTL(time.Minute),
			evcache.WithPolicy(evcache.LRU),
		)

		h := NewEventsHandler(db, tagsCache, eventsCache)
		h.Register(g)
	}

	{
		g := g.Group("",
			echo.WrapMiddleware(NoCache),
		)

		h := NewFeedHandler(db, baseURL, eventsCache)
		h.Register(g)
	}
}
