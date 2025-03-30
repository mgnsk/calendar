package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/wreck"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// UsersHandler handles users pages.
type UsersHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Users handles users page.
func (h *UsersHandler) Users(c *Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	users, err := model.ListUsers(c.Request().Context(), h.db)
	if err != nil {
		return err
	}

	return RenderPage(c, h.sm,
		html.UsersMain(c.User, users, c.CSRF),
	)
}

// Invite handles invite link generation.
func (h *UsersHandler) Invite(c *Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	if c.User.Role != domain.Admin {
		return wreck.Forbidden.New("Only admins can invite users")
	}

	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		token := uuid.New()

		if err := model.InsertInvite(c.Request().Context(), h.db, &domain.Invite{
			Token:      token,
			ValidUntil: time.Now().Add(72 * time.Hour),
			CreatedBy:  c.User.ID,
		}); err != nil {
			return err

		}

		return html.InviteLinkPartial(token).Render(c.Response())
	}

	return wreck.NotFound.New("Not found")
}

// RegisterUser registers a user with an invite link.
func (h *UsersHandler) RegisterUser(c *Context) error {
	if c.User != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	req := contract.RegisterRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	invite, err := model.GetInvite(c.Request().Context(), h.db, req.Token)
	if err != nil {
		return err
	}

	if !invite.IsValid() {
		return wreck.NotFound.New("Not found")
	}

	switch c.Request().Method {
	case http.MethodGet:
		form := contract.RegisterForm{}

		return RenderPage(c, h.sm,
			html.RegisterMain(form, nil, c.CSRF),
		)

	case http.MethodPost:
		form := contract.RegisterForm{}
		if err := c.Bind(&form); err != nil {
			return err
		}

		if errs := form.Validate(); len(errs) > 0 {
			return RenderPage(c, h.sm,
				html.RegisterMain(form, errs, c.CSRF),
			)
		}

		newUser := &domain.User{
			ID:       snowflake.Generate(),
			Username: form.Username,
			Role:     domain.Author,
		}

		if err := newUser.SetPassword(form.Password1); err != nil {
			if errors.Is(err, wreck.InvalidValue) {
				errs := url.Values{}
				errs.Set("password1", err.Error())
				errs.Set("password2", err.Error())

				return RenderPage(c, h.sm,
					html.RegisterMain(form, errs, c.CSRF),
				)
			}

			return err
		}

		if err := h.db.RunInTx(c.Request().Context(), nil, func(ctx context.Context, tx bun.Tx) error {
			if err := model.DeleteInvite(ctx, tx, invite.Token); err != nil {
				return err
			}

			return model.InsertUser(ctx, tx, newUser)
		}); err != nil {
			if errors.Is(err, wreck.AlreadyExists) {
				errs := url.Values{}
				errs.Set("username", "User already exists")

				return RenderPage(c, h.sm,
					html.RegisterMain(form, errs, c.CSRF),
				)
			}

			return err
		}

		// First renew the session token.
		if err := h.sm.RenewToken(c.Request().Context()); err != nil {
			return err
		}

		// Then make the privilege-level change.
		h.sm.Put(c.Request().Context(), "username", newUser.Username)

		return c.Redirect(http.StatusSeeOther, "/")

	default:
		return wreck.NotFound.New("Not found")
	}
}

// Delete a user.
func (h *UsersHandler) Delete(c *Context) error {
	if c.User == nil {
		return wreck.Forbidden.New("Must be logged in")
	}

	if c.User.Role != domain.Admin {
		return wreck.Forbidden.New("Only admins can delete users")
	}

	req := contract.DeleteUserRequest{}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if c.User.Username == req.Username {
		return wreck.Forbidden.New("Cannot delete yourself")
	}

	if c.Request().Method == http.MethodPost && hxhttp.IsRequest(c.Request().Header) {
		if err := model.DeleteUser(c.Request().Context(), h.db, req.Username); err != nil {
			return err
		}

		h.sm.Put(c.Request().Context(), "flash-success", fmt.Sprintf("User %s deleted", req.Username))

		hxhttp.SetRefresh(c.Response().Header())

		return nil
	}

	return wreck.NotFound.New("Not found")
}

// Register the handler.
func (h *UsersHandler) Register(g *echo.Group) {
	g.GET("/users", Wrap(h.db, h.sm, h.Users))

	g.POST("/delete-user", Wrap(h.db, h.sm, h.Delete))
	g.POST("/invite", Wrap(h.db, h.sm, h.Invite))

	g.GET("/register/:token", Wrap(h.db, h.sm, h.RegisterUser))
	g.POST("/register/:token", Wrap(h.db, h.sm, h.RegisterUser))
}

// NewUsersHandler creates a new users handler.
func NewUsersHandler(db *bun.DB, sm *scs.SessionManager) *UsersHandler {
	return &UsersHandler{
		db: db,
		sm: sm,
	}
}
