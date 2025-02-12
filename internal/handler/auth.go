package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// AuthenticationHandler handles user login and logout.
type AuthenticationHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Login handles login page.
func (h *AuthenticationHandler) Login(c echo.Context) error {
	user := loadUser(c)
	if user != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	settings := loadSettings(c)

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.LoginPage(settings.Title, nil, c.Get("csrf").(string), "", "").Render(c.Response())

	case http.MethodPost:
		username := c.FormValue("username")
		password := c.FormValue("password")

		invalidLogin := func() error {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			errs := map[string]string{
				"username": "Invalid username or password",
				"password": "Invalid username or password",
			}

			return html.LoginPage(settings.Title, errs, c.Get("csrf").(string), username, password).Render(c.Response())
		}

		if username == "" || password == "" {
			return invalidLogin()
		}

		// Grace timeout for login failures so we always fail in constant time
		// regardless of whether user does not exist or invalid password provided.
		ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
		defer cancel()

		user, err := model.GetUser(ctx, h.db, username)
		if err != nil {
			if errors.Is(err, wreck.NotFound) {
				<-ctx.Done()
				return invalidLogin()
			}
			return err
		}

		if err := user.VerifyPassword(password); err != nil {
			if errors.Is(err, wreck.InvalidValue) {
				<-ctx.Done()
				return invalidLogin()
			}
			return err
		}

		// First renew the session token.
		if err := h.sm.RenewToken(c.Request().Context()); err != nil {
			return err
		}

		// Then make the privilege-level change.
		h.sm.Put(c.Request().Context(), "username", user.Username)

		return c.Redirect(http.StatusSeeOther, "/")

	default:
		panic("unhandled method")
	}
}

// Login handles logout page.
func (h *AuthenticationHandler) Logout(c echo.Context) error {
	if err := h.sm.Destroy(c.Request().Context()); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

// Register the handler.
func (h *AuthenticationHandler) Register(g *echo.Group) {
	g.GET("/login", h.Login)
	g.POST("/login", h.Login)

	g.GET("/logout", h.Logout)
}

// NewAuthenticationHandler creates a new authentication handler.
func NewAuthenticationHandler(db *bun.DB, sm *scs.SessionManager) *AuthenticationHandler {
	return &AuthenticationHandler{
		db: db,
		sm: sm,
	}
}
