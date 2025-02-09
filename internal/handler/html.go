package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// LimitPerPage specifies the maximum numbers of events per page.
const LimitPerPage = 3

// HTMLHandler handles web pages.
type HTMLHandler struct {
	db      *bun.DB
	sm      *scs.SessionManager
	baseURL *url.URL
}

func (h *HTMLHandler) getTagFilter(c echo.Context) (string, error) {
	if param := c.Param("tagName"); param != "" {
		v, err := url.QueryUnescape(param)
		if err != nil {
			return "", wreck.InvalidValue.New("Invalid tag filter", err)
		}
		return v, nil
	}

	return "", nil
}

func (h *HTMLHandler) getIntParam(key string, c echo.Context) (int64, error) {
	if v := c.FormValue(key); v != "" {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, wreck.InvalidValue.New(fmt.Sprintf("Invalid %s", key), err)
		}
		return val, nil
	}

	return 0, nil
}

func (h *HTMLHandler) getOffset(c echo.Context) (int64, error) {
	var offset int64
	if v := c.FormValue("offset"); v != "" {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, wreck.InvalidValue.New("Invalid offset", err)
		}
		offset = val + LimitPerPage
		return offset, nil
	}

	return 0, nil
}

// Setup handles the setup page.
func (h *HTMLHandler) Setup(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		var s *domain.Settings
		if v := c.Get("settings"); v != nil {
			s = v.(*domain.Settings)
		} else {
			s = domain.NewDefaultSettings()
		}

		vals := url.Values{}
		vals.Set("title", s.Title)
		vals.Set("description", s.Description)

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

		s := domain.NewDefaultSettings()
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

// Upcoming handles the upcoming events page.
func (h *HTMLHandler) Upcoming(c echo.Context) error {
	filterTag, err := h.getTagFilter(c)
	if err != nil {
		return err
	}

	offset, err := h.getOffset(c)
	if err != nil {
		return err
	}

	events, err := model.NewEventsQuery().
		WithStartAtFrom(time.Now()).
		WithOrder(offset, model.OrderStartAtAsc).
		WithLimit(LimitPerPage).
		WithFilterTags(filterTag).
		List(c.Request().Context(), h.db, c.FormValue("search"))
	if err != nil {
		if !errors.Is(err, wreck.NotFound) {
			return err
		}
	}

	if hxhttp.IsRequest(c.Request().Header) {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(offset, events, c.Get("csrf").(string), c.Path()).Render(c.Response())
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	s := c.Get("settings").(*domain.Settings)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    s.Title,
		SectionTitle: "Upcoming events",
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		Offset:       offset,
		Events:       events,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

// PastEvents handles past events page.
func (h *HTMLHandler) PastEvents(c echo.Context) error {
	filterTag, err := h.getTagFilter(c)
	if err != nil {
		return err
	}

	offset, err := h.getOffset(c)
	if err != nil {
		return err
	}

	// Lists events that have already started, in descending order.
	events, err := model.NewEventsQuery().
		WithStartAtUntil(time.Now()).
		WithOrder(offset, model.OrderStartAtDesc).
		WithLimit(LimitPerPage).
		WithFilterTags(filterTag).
		List(c.Request().Context(), h.db, c.FormValue("search"))
	if err != nil {
		if !errors.Is(err, wreck.NotFound) {
			return err
		}
	}

	if hxhttp.IsRequest(c.Request().Header) {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(offset, events, c.Get("csrf").(string), c.Path()).Render(c.Response())
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	s := c.Get("settings").(*domain.Settings)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    s.Title,
		SectionTitle: "Past Events",
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		Offset:       offset,
		Events:       events,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

// LatestEvents returns latest events.
func (h *HTMLHandler) LatestEvents(c echo.Context) error {
	filterTag, err := h.getTagFilter(c)
	if err != nil {
		return err
	}

	cursor, err := h.getIntParam("last_id", c)
	if err != nil {
		return err
	}

	// Latest events sorted in newest created first.
	// Past events are filtered out.
	events, err := model.NewEventsQuery().
		WithStartAtFrom(time.Now()).
		WithOrder(cursor, model.OrderCreatedAtDesc).
		WithLimit(LimitPerPage).
		WithFilterTags(filterTag).
		List(c.Request().Context(), h.db, c.FormValue("search"))
	if err != nil {
		if !errors.Is(err, wreck.NotFound) {
			return err
		}
	}

	if hxhttp.IsRequest(c.Request().Header) {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(0, events, c.Get("csrf").(string), c.Path()).Render(c.Response())
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	s := c.Get("settings").(*domain.Settings)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    s.Title,
		SectionTitle: "Latest Events",
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		Offset:       0,
		Events:       events,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

// Tags handles tag list page.
func (h *HTMLHandler) Tags(c echo.Context) error {
	tags, err := model.ListTags(c.Request().Context(), h.db)
	if err != nil {
		return err
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	s := c.Get("settings").(*domain.Settings)

	return html.TagsPage(html.TagsPageParams{
		MainTitle:    s.Title,
		SectionTitle: "Tags",
		Path:         c.Path(),
		User:         user,
		Tags:         tags,
		CSRF:         c.Get("csrf").(string),
	}).Render(c.Response())
}

// Manage handles management.
func (h *HTMLHandler) Manage(c echo.Context) error {
	panic("not implemented")
}

// Login handles user login.
func (h *HTMLHandler) Login(c echo.Context) error {
	if currentUser, err := h.loadUser(c); err != nil {
		return err
	} else if currentUser != nil {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	switch c.Request().Method {
	case http.MethodGet:
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)

		s := c.Get("settings").(*domain.Settings)

		return html.LoginPage(s.Title, nil, c.Get("csrf").(string), "", "").Render(c.Response())

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

			s := c.Get("settings").(*domain.Settings)

			return html.LoginPage(s.Title, errs, c.Get("csrf").(string), username, password).Render(c.Response())
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
			<-ctx.Done()
			return invalidLogin()
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

// Logout handles user logout.
func (h *HTMLHandler) Logout(c echo.Context) error {
	if err := h.sm.Destroy(c.Request().Context()); err != nil {
		return err
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *HTMLHandler) loadUser(c echo.Context) (*domain.User, error) {
	username := h.sm.GetString(c.Request().Context(), "username")
	if username == "" {
		return nil, nil
	}

	user, err := model.GetUser(c.Request().Context(), h.db, username)
	if err != nil {
		if errors.Is(err, wreck.NotFound) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

// Register the handler.
func (h *HTMLHandler) Register(e *echo.Echo) {
	h.sm.ErrorFunc = func(w http.ResponseWriter, r *http.Request, err error) {
		if err := HandleError(err, e.NewContext(r, w)); err != nil {
			panic(err)
		}
	}

	// Serve assets from the embedded filesystem.
	e.GET("/dist/*",
		echo.StaticDirectoryHandler(echo.MustSubFS(internal.DistFS, "dist"), false),
		AssetCacheMiddleware,
	)

	g := e.Group("",
		LoadSettingsMiddleware(h.db),
		echo.WrapMiddleware(h.sm.LoadAndSave),
	)

	g.GET("/", h.LatestEvents)
	g.POST("/", h.LatestEvents) // Fox htmx.
	g.GET("/tag/:tagName", h.LatestEvents)
	g.POST("/tag/:tagName", h.LatestEvents) // For htmx.

	g.GET("/upcoming", h.Upcoming)
	g.POST("/upcoming", h.Upcoming) // Fox htmx.
	g.GET("/upcoming/tag/:tagName", h.Upcoming)
	g.POST("/upcoming/tag/:tagName", h.Upcoming) // For htmx.

	g.GET("/past", h.PastEvents)
	g.POST("/past", h.PastEvents) // For htmx.
	g.GET("/past/tag/:tagName", h.PastEvents)
	g.POST("/past/tag/:tagName", h.PastEvents) // For htmx.

	g.GET("/tags", h.Tags)

	g.GET("/setup", h.Setup, echo.WrapMiddleware(NoCache))
	g.POST("/setup", h.Setup, echo.WrapMiddleware(NoCache))

	g.GET("/login", h.Login, echo.WrapMiddleware(NoCache))
	g.POST("/login", h.Login, echo.WrapMiddleware(NoCache))

	g.GET("/logout", h.Logout, echo.WrapMiddleware(NoCache))

	g.GET("/manage", h.Manage, echo.WrapMiddleware(NoCache))
	g.POST("/manage", h.Manage, echo.WrapMiddleware(NoCache))
}

// NewHTMLHandler creates a new HTML handler.
func NewHTMLHandler(db *bun.DB, store scs.Store, baseURL *url.URL) *HTMLHandler {
	// Initialize a new session manager and configure the session lifetime.
	sm := scs.New()
	sm.Store = store
	sm.HashTokenInStore = true
	sm.Lifetime = 24 * time.Hour
	// sm.IdleTimeout = 20 * time.Minute // TODO
	sm.Cookie.Name = "session_id"
	sm.Cookie.Domain = ""
	sm.Cookie.HttpOnly = true
	sm.Cookie.Path = "/"
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode
	sm.Cookie.Secure = true

	return &HTMLHandler{
		db:      db,
		sm:      sm,
		baseURL: baseURL,
	}
}
