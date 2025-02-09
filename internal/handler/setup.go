package handler

import (
	"context"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
)

// SetupHandler handles setup pages.
type SetupHandler struct {
	db *bun.DB
}

// Setup handles the setup page.
func (h *SetupHandler) Setup(c echo.Context) error {
	settings := loadSettings(c)
	if settings != nil {
		// Already set up.
		return wreck.NotFound.New("")
	} else {
		settings = domain.NewDefaultSettings()
	}

	switch c.Request().Method {
	case http.MethodGet:

		vals := url.Values{}
		vals.Set("title", settings.Title)
		vals.Set("description", settings.Description)

		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		return html.SetupPage(vals, nil, c.Get("csrf").(string)).Render(c.Response())

	case http.MethodPost:
		// TODO: Implement some form validation framework
		errs := map[string]string{}

		title := c.FormValue("title")
		if title == "" {
			errs["title"] = "Title must be set"
		}

		desc := c.FormValue("description")
		if title == "" {
			errs["description"] = "Description must be set"
		}

		username := c.FormValue("username")
		if username == "" {
			errs["username"] = "Username must be set"
		}

		password1 := c.FormValue("password1")
		if password1 == "" {
			errs["password1"] = "Password must be set"
		}

		password2 := c.FormValue("password2")
		if password2 == "" {
			errs["password2"] = "Password must be set"
		}

		if password1 != password2 {
			errs["password2"] = "Passwords must match"
		}

		if len(errs) > 0 {
			c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
			c.Response().WriteHeader(200)

			form, err := c.FormParams()
			if err != nil {
				return err
			}

			return html.SetupPage(form, errs, c.Get("csrf").(string)).Render(c.Response())
		}

		settings.Title = title
		settings.Description = desc

		if err := h.db.RunInTx(c.Request().Context(), nil, func(ctx context.Context, tx bun.Tx) error {
			if err := model.InsertSettings(ctx, tx, settings); err != nil {
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

		// TODO: currently the user must explicitly log in after setup
		// // First renew the session token.
		// if err := h.sm.RenewToken(c.Request().Context()); err != nil {
		// 	return err
		// }
		//
		// // Then make the privilege-level change.
		// h.sm.Put(c.Request().Context(), "username", username)

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
func NewSetupHandler(db *bun.DB) *SetupHandler {
	return &SetupHandler{
		db: db,
	}
}
