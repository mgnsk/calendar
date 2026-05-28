package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/ggicci/httpin"
	"github.com/google/uuid"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/server"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// UsersHandler handles users pages.
type UsersHandler struct {
	db *bun.DB
	sm *scs.SessionManager
}

// Users handles users page.
func (h *UsersHandler) Users(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	if user.Role != domain.Admin {
		panic(calendar.Forbidden.New("Only admins can view users"))
	}

	users, err := model.ListUsers(r.Context(), h.db)
	if err != nil {
		panic(err)
	}

	server.RenderPage(w, r, h.sm,
		html.UsersMain(user, users),
	)
}

// Invite handles invite link generation.
func (h *UsersHandler) Invite(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	if user.Role != domain.Admin {
		panic(calendar.Forbidden.New("Only admins can invite users"))
	}

	if r.Method == http.MethodPost && hxhttp.IsRequest(r.Header) {
		token := uuid.New()

		if err := model.InsertInvite(r.Context(), h.db, &domain.Invite{
			Token:      token,
			ValidUntil: time.Now().Add(72 * time.Hour),
			CreatedBy:  user.ID,
		}); err != nil {
			panic(err)

		}

		if err := html.InviteLinkPartial(token).Render(w); err != nil {
			panic(err)
		}
		return
	}

	panic(calendar.NotFound.New("Not found"))
}

// RegisterUser registers a user with an invite link.
func (h *UsersHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	req := contract.RegisterRequest{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	invite, err := model.GetInvite(r.Context(), h.db, req.Token)
	if err != nil {
		panic(err)
	}

	if !invite.IsValid() {
		panic(calendar.NotFound.New("Not found"))
	}

	switch r.Method {
	case http.MethodGet:
		form := contract.RegisterForm{}

		server.RenderPage(w, r, h.sm,
			html.RegisterMain(form, nil),
		)
		return

	case http.MethodPost:
		form := contract.RegisterForm{}
		if err := httpin.DecodeTo(r, &form); err != nil {
			panic(err)
		}

		if errs := form.Validate(); len(errs) > 0 {
			server.RenderPage(w, r, h.sm,
				html.RegisterMain(form, errs),
			)
			return
		}

		newUser := &domain.User{
			ID:       snowflake.Generate(),
			Username: form.Username,
			Role:     domain.Author,
		}

		if err := newUser.SetPassword(form.Password1); err != nil {
			if errors.Is(err, calendar.InvalidValue) {
				errs := url.Values{}
				errs.Set("password1", err.Error())
				errs.Set("password2", err.Error())

				server.RenderPage(w, r, h.sm,
					html.RegisterMain(form, errs),
				)
				return
			}

			panic(err)
		}

		if err := h.db.RunInTx(r.Context(), nil, func(ctx context.Context, db bun.Tx) error {
			if err := model.DeleteInvite(ctx, db, invite.Token); err != nil {
				return err
			}

			return model.InsertUser(ctx, db, newUser)
		}); err != nil {
			if errors.Is(err, calendar.AlreadyExists) {
				errs := url.Values{}
				errs.Set("username", "User already exists")

				server.RenderPage(w, r, h.sm,
					html.RegisterMain(form, errs),
				)
				return
			}

			panic(err)
		}

		// First renew the session token.
		if err := h.sm.RenewToken(r.Context()); err != nil {
			panic(err)
		}

		// Then make the privilege-level change.
		h.sm.Put(r.Context(), "username", newUser.Username)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return

	default:
		panic(calendar.NotFound.New("Not found"))
	}
}

// Delete a user.
func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	if user.Role != domain.Admin {
		panic(calendar.Forbidden.New("Only admins can delete users"))
	}

	req := contract.DeleteUserRequest{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	if user.ID == req.UserID {
		panic(calendar.Forbidden.New("Cannot delete yourself"))
	}

	if r.Method == http.MethodPost && hxhttp.IsRequest(r.Header) {
		if err := model.DeleteUser(r.Context(), h.db, req.UserID); err != nil {
			panic(err)
		}

		h.sm.Put(r.Context(), "flash-success", "User deleted")

		hxhttp.SetRefresh(w.Header())

		return
	}

	panic(calendar.NotFound.New("Not found"))
}

// UpgradeUserRole upgrades user role.
func (h *UsersHandler) UpgradeUserRole(w http.ResponseWriter, r *http.Request) {
	user := server.GetUser(r.Context())
	if user == nil {
		panic(calendar.Forbidden.New("Must be logged in"))
	}

	if user.Role != domain.Admin {
		panic(calendar.Forbidden.New("Only admins can upgrade users"))
	}

	req := contract.UpgradeUserRoleRequest{}
	if err := httpin.DecodeTo(r, &req); err != nil {
		panic(err)
	}

	if user.ID == req.UserID {
		panic(calendar.Forbidden.New("Cannot upgrade yourself"))
	}

	if r.Method == http.MethodPost && hxhttp.IsRequest(r.Header) {
		user, err := model.GetUser(r.Context(), h.db, req.UserID)
		if err != nil {
			panic(err)
		}

		user.Role = domain.Admin

		if err := model.UpdateUser(r.Context(), h.db, user); err != nil {
			panic(err)
		}

		h.sm.Put(r.Context(), "flash-success", "User upgraded to admin")

		hxhttp.SetRefresh(w.Header())

		return
	}

	panic(calendar.NotFound.New("Not found"))
}

// Register the handler.
func (h *UsersHandler) Register(mux *http.ServeMux) {
	middlewares := []server.MiddlewareFunc{
		server.NewSessionMiddleware(h.sm),
		server.NewSettingsMiddleware(h.db),
		server.NewUserMiddleware(h.db, h.sm),
	}

	mux.Handle("GET /users", server.WithMiddleware(http.HandlerFunc(h.Users), middlewares...))

	mux.Handle("POST /delete-user", server.WithMiddleware(http.HandlerFunc(h.Delete), middlewares...))
	mux.Handle("POST /upgrade-user", server.WithMiddleware(http.HandlerFunc(h.UpgradeUserRole), middlewares...))
	mux.Handle("POST /invite", server.WithMiddleware(http.HandlerFunc(h.Invite), middlewares...))

	mux.Handle("GET /register/{token}", server.WithMiddleware(http.HandlerFunc(h.RegisterUser), middlewares...))
	mux.Handle("POST /register/{token}", server.WithMiddleware(http.HandlerFunc(h.RegisterUser), middlewares...))
}

// NewUsersHandler creates a new users handler.
func NewUsersHandler(db *bun.DB, sm *scs.SessionManager) *UsersHandler {
	return &UsersHandler{
		db: db,
		sm: sm,
	}
}
