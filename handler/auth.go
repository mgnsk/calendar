package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/wreck"
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

	s := loadSettings(c)
	csrf := c.Get("csrf").(string)

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.Page(html.PageProps{
			Title:        s.Title,
			User:         user,
			Path:         c.Path(),
			CSRF:         csrf,
			Children:     html.LoginMain(contract.LoginForm{}, nil, csrf),
			FlashSuccess: "",
		}).Render(c.Response())

	case http.MethodPost:
		req := contract.LoginForm{}
		if err := c.Bind(&req); err != nil {
			return err
		}

		invalidLogin := func(errs url.Values) error {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(html.PageProps{
				Title:        s.Title,
				User:         user,
				Path:         c.Path(),
				CSRF:         csrf,
				Children:     html.LoginMain(contract.LoginForm{}, errs, csrf),
				FlashSuccess: "",
			}).Render(c.Response())
		}

		if errs := req.Validate(); len(errs) > 0 {
			return invalidLogin(errs)
		}

		// Grace timeout for login failures so we always fail in constant time
		// regardless of whether user does not exist or invalid password provided.
		ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
		defer cancel()

		user, err := model.GetUser(ctx, h.db, req.Username)
		if err != nil {
			if errors.Is(err, wreck.NotFound) {
				<-ctx.Done()
				errs := url.Values{}
				errs.Set("username", "Invalid username or password")
				errs.Set("password", "Invalid username or password")
				return invalidLogin(errs)
			}
			return err
		}

		if err := user.VerifyPassword(req.Password); err != nil {
			if errors.Is(err, wreck.InvalidValue) {
				<-ctx.Done()
				errs := url.Values{}
				errs.Set("username", "Invalid username or password")
				errs.Set("password", "Invalid username or password")
				return invalidLogin(errs)
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

// Logout handles logout page.
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
