package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/contract"
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
	rc := GetContext(c)
	if rc.Settings != nil {
		// Already set up.
		return wreck.NotFound.New("")
	}

	rc.Settings = domain.NewDefaultSettings()

	switch c.Request().Method {
	case http.MethodGet:
		form := contract.SetupForm{
			Title:       rc.Settings.Title,
			Description: rc.Settings.Description,
		}

		return RenderPage(c, rc,
			html.SetupMain(form, nil, rc.CSRF),
		)

	case http.MethodPost:
		form := contract.SetupForm{}
		if err := c.Bind(&form); err != nil {
			return err
		}

		if errs := form.Validate(); len(errs) > 0 {
			return RenderPage(c, rc,
				html.SetupMain(form, errs, rc.CSRF),
			)
		}

		rc.Settings.Title = form.Title
		rc.Settings.Description = form.Description

		user := &domain.User{
			ID:       snowflake.Generate(),
			Username: form.Username,
			Role:     domain.Admin,
		}

		if err := user.SetPassword(form.Password1); err != nil {
			if errors.Is(err, wreck.InvalidValue) {
				errs := url.Values{}
				errs.Set("password1", err.Error())
				errs.Set("password2", err.Error())

				return RenderPage(c, rc,
					html.SetupMain(form, errs, rc.CSRF),
				)
			}

			return err
		}

		if err := h.db.RunInTx(c.Request().Context(), nil, func(ctx context.Context, tx bun.Tx) error {
			if err := model.InsertSettings(ctx, tx, rc.Settings); err != nil {
				return err
			}

			return model.InsertUser(ctx, tx, user)
		}); err != nil {
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
