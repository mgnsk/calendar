package handler

import (
	"context"
	"net/http"
	"net/url"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/uptrace/bun"
)

// SetupHandler handles setup pages.
type SetupHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Setup handles the setup page.
func (h *SetupHandler) Setup(c echo.Context) error {
	s := loadSettings(c)
	if s != nil {
		// Already set up.
		return wreck.NotFound.New("")
	}

	s = domain.NewDefaultSettings()
	csrf := c.Get("csrf").(string)

	switch c.Request().Method {
	case http.MethodGet:
		form := url.Values{}
		form.Set("pagetitle", s.Title)
		form.Set("pagedesc", s.Description)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.Page(s.Title, nil, c.Path(), csrf, html.SetupMain(form, nil, csrf)).Render(c.Response())

	case http.MethodPost:
		form, err := c.FormParams()
		if err != nil {
			return err
		}
		errs := url.Values{}

		title := c.FormValue("pagetitle")
		if title == "" {
			errs.Set("pagetitle", "Title must be set")
		}

		desc := c.FormValue("pagedesc")
		if desc == "" {
			errs.Set("pagedesc", "Description must be set")
		}

		username := c.FormValue("username")
		if username == "" {
			errs.Set("username", "Username must be set")
		}

		password1 := c.FormValue("password1")
		if password1 == "" {
			errs.Set("password1", "Password must be set")
		}

		password2 := c.FormValue("password2")
		if password2 == "" {
			errs.Set("password2", "Password must be set")
		}

		if password1 != password2 {
			errs.Set("password2", "Passwords must match")
		}

		if len(errs) > 0 {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			return html.Page(s.Title, nil, c.Path(), csrf, html.SetupMain(form, errs, csrf)).Render(c.Response())
		}

		s.Title = title
		s.Description = desc

		if err := h.db.RunInTx(c.Request().Context(), nil, func(ctx context.Context, tx bun.Tx) error {
			if err := model.InsertSettings(ctx, tx, s); err != nil {
				return err
			}

			user := &domain.User{
				ID:       snowflake.Generate(),
				Username: username,
				Role:     domain.Admin,
			}
			user.SetPassword(password1)

			return model.InsertUser(ctx, tx, user)
		}); err != nil {
			return err
		}

		// First renew the session token.
		if err := h.sm.RenewToken(c.Request().Context()); err != nil {
			return err
		}

		// Then make the privilege-level change.
		h.sm.Put(c.Request().Context(), "username", username)

		return c.Redirect(http.StatusSeeOther, "/")

	default:
		panic("unhandled method")
	}
}

// Register the handler.
func (h *SetupHandler) Register(g *echo.Group) {
	g.GET("/setup", h.Setup)
	g.POST("/setup", h.Setup)
}

// NewSetupHandler creates a new setup handler.
func NewSetupHandler(db *bun.DB, sm *scs.SessionManager) *SetupHandler {
	return &SetupHandler{
		db: db,
		sm: sm,
	}
}
