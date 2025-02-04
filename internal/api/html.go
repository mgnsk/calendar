package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"github.com/uptrace/bun"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

// TODO: error handling, bad request, not found, etc

// LimitPerPage specifies the maximum numbers of events per page.
const LimitPerPage = 3

// HTMLHandler handles web pages.
type HTMLHandler struct {
	db     *bun.DB
	config Config
}

// Register the handler.
func (h *HTMLHandler) Register(e *echo.Echo) {
	// Serve assets from the embedded filesystem.
	e.GET("/dist/*",
		echo.StaticDirectoryHandler(echo.MustSubFS(internal.DistFS, "dist"), false),
		AssetCacheMiddleware,
	)

	g := e.Group("",
		session.Middleware(sessions.NewCookieStore(h.config.SessionSecret)),
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

	// g.GET("/setup", h.Setup, echo.WrapMiddleware(NoCache))
	// g.POST("/setup", h.Setup, echo.WrapMiddleware(NoCache))

	g.GET("/login", h.Login, echo.WrapMiddleware(NoCache))
	g.POST("/login", h.Login, echo.WrapMiddleware(NoCache))

	g.GET("/logout", h.Logout, echo.WrapMiddleware(NoCache))

	g.GET("/manage", h.Manage, echo.WrapMiddleware(NoCache))
	g.POST("/manage", h.Manage, echo.WrapMiddleware(NoCache))
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

// // Setup handles the setup page.
// func (h *HTMLHandler) Setup(c echo.Context) error {
// 	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
// 	c.Response().WriteHeader(200)
//
// 	return html.SetupPage(false, c.Get("settings").(*domain.Settings)).Render(c.Response())
// }

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

	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now(), time.Time{}, c.FormValue("search"), model.OrderStartAtAsc, offset, LimitPerPage, filterTag)
	if err != nil {
		if !errors.Is(err, wreck.NotFound) {
			return err
		}
	}

	if hxhttp.IsRequest(c.Request().Header) {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(offset, events, c.Path()).Render(c.Response())
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    h.config.PageTitle,
		SectionTitle: "Upcoming events",
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		Offset:       offset,
		Events:       events,
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
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Time{}, time.Now(), c.FormValue("search"), model.OrderStartAtDesc, offset, LimitPerPage, filterTag)
	if err != nil {
		if !errors.Is(err, wreck.NotFound) {
			return err
		}
	}

	if hxhttp.IsRequest(c.Request().Header) {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(offset, events, c.Path()).Render(c.Response())
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    h.config.PageTitle,
		SectionTitle: "Past Events",
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		Offset:       offset,
		Events:       events,
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
	events, err := model.ListEvents(c.Request().Context(), h.db, time.Now(), time.Time{}, c.FormValue("search"), model.OrderCreatedAtDesc, cursor, LimitPerPage, filterTag)
	if err != nil {
		if !errors.Is(err, wreck.NotFound) {
			return err
		}
	}

	if hxhttp.IsRequest(c.Request().Header) {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(200)
		return html.EventListPartial(0, events, c.Path()).Render(c.Response())
	}

	user, err := h.loadUser(c)
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(200)

	return html.EventsPage(html.EventsPageParams{
		MainTitle:    h.config.PageTitle,
		SectionTitle: "Latest Events",
		Path:         c.Path(),
		FilterTag:    filterTag,
		User:         user,
		Offset:       0,
		Events:       events,
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

	return html.TagsPage(html.TagsPageParams{
		MainTitle:    h.config.PageTitle,
		SectionTitle: "Tags",
		Path:         c.Path(),
		User:         user,
		Tags:         tags,
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
			// Grace timeout for login failures so we always fail in constant time
			// regardless of whether user does not exist or invalid password provided.
			ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
			defer cancel()

			user, err := model.GetUser(ctx, h.db, username)
			if err != nil {
				if errors.Is(err, wreck.NotFound) {
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
		if errors.Is(err, wreck.NotFound) {
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
