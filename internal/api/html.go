package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// HTMLHandler handles web pages.
type HTMLHandler struct {
	db     *bun.DB
	config Config
}

// NoCacheMiddleware disables caching for responses.
func NoCacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // HTTP 1.1.
		c.Response().Header().Set("Pragma", "no-cache")                                   // HTTP 1.0.
		c.Response().Header().Set("Expires", "0")                                         // Proxies.

		return next(c)
	}
}

// Register the handler.
func (h *HTMLHandler) Register(e *echo.Echo) {
	g := e.Group("",
		session.Middleware(sessions.NewCookieStore(h.config.SessionSecret)),
	)

	g.GET("/", h.Home)
	g.GET("/past", h.PastEvents)

	g.GET("/login", h.Login, NoCacheMiddleware)
	g.POST("/login", h.Login, NoCacheMiddleware)

	g.GET("/logout", h.Logout, NoCacheMiddleware)

	g.GET("/users", h.Users, NoCacheMiddleware)
	g.POST("/users", h.Users, NoCacheMiddleware)

	g.GET("/change-password", h.ChangePassword, NoCacheMiddleware)
	g.POST("/change-password", h.ChangePassword, NoCacheMiddleware)
}

// Home handles the home page.
func (h *HTMLHandler) Home(c echo.Context) error {
	// Lists events that started in the past 24 hours, start time ascending.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now().Add(-24*time.Hour), time.Time{}, "asc")
	if err != nil {
		return err
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	return html.CurrentEventsPage(h.config.PageTitle, user, events).Render(c.Response())
}

// PastEvents handles past events page.
func (h *HTMLHandler) PastEvents(c echo.Context) error {
	// Lists events that have already started, in descending order.
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Time{}, time.Now(), "desc")
	if err != nil {
		return err
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	return html.PastEventsPage(h.config.PageTitle, user, events).Render(c.Response())
}

// Users handles managing users.
func (h *HTMLHandler) Users(c echo.Context) error {
	panic("not implemented")
}

// ChangePassword handles changing user's password.
func (h *HTMLHandler) ChangePassword(c echo.Context) error {
	panic("not implemented")
}

// Login handles user login.
func (h *HTMLHandler) Login(c echo.Context) error {
	if currentUser, err := h.loadUser(c); err != nil {
		return err
	} else if currentUser != nil {
		return c.Redirect(http.StatusFound, "/")
	}

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.LoginPage(h.config.PageTitle, false, "", "").Render(c.Response())

	case http.MethodPost:
		username := c.FormValue("username")
		password := c.FormValue("password")
		if username == "" || password == "" {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.LoginPage(h.config.PageTitle, true, username, password).Render(c.Response())
		}

		{
			// Grace timeout for login failures so we always fail in constant time.
			ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
			defer cancel()

			user, err := model.GetUser(ctx, h.db, username)
			if err != nil {
				if e := new(wreck.NotFound); errors.As(err, &e) {
					<-ctx.Done()

					c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
					c.Response().WriteHeader(200)

					return html.LoginPage(h.config.PageTitle, true, username, password).Render(c.Response())
				}

				return err
			}

			if err := user.VerifyPassword(password); err != nil {
				<-ctx.Done()

				c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
				c.Response().WriteHeader(200)

				return html.LoginPage(h.config.PageTitle, true, username, password).Render(c.Response())
			}
		}

		sess, err := session.Get("session", c)
		if err != nil {
			return err
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 7,
			HttpOnly: true,
			Secure:   false, // TODO: upgrade this when running on https
			// TODO: cookie actually set to SameSite=None by default
			// SameSite: http.SameSiteNoneMode,
		}

		sess.Values["username"] = username
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		return c.Redirect(http.StatusFound, "/")

	default:
		panic("unhandled method")
	}
}

// Logout handles user logout.
func (*HTMLHandler) Logout(c echo.Context) error {
	sess, err := session.Get("session", c)
	if err != nil {
		return err
	}

	sess.Values = nil
	sess.Options.MaxAge = -1

	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/")
}

func (h *HTMLHandler) loadUser(c echo.Context) (*domain.User, error) {
	sess, err := session.Get("session", c)
	if err != nil {
		return nil, err
	}

	username, ok := sess.Values["username"].(string)
	if !ok || username == "" {
		return nil, nil
	}

	user, err := model.GetUser(c.Request().Context(), h.db, username)
	if err != nil {
		if e := new(wreck.NotFound); errors.As(err, &e) {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

// NewHTMLHandler creates a new HTML handler.
func NewHTMLHandler(db *bun.DB, config Config) *HTMLHandler {
	return &HTMLHandler{
		db:     db,
		config: config,
	}
}
