package handler_test

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/alexedwards/scs/bunstore"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/handler"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeTimezoneFinder is a stub implementing handler.TimezoneFinder, avoiding
// a dependency on the real (heavy) tzf geo timezone database in tests.
type fakeTimezoneFinder struct {
	// name is the IANA timezone name to always return. An empty name makes
	// EditEventHandler fall back to the submitted user_timezone form field,
	// then to UTC.
	name string
}

func (f fakeTimezoneFinder) GetTimezoneName(_, _ float64) string {
	return f.name
}

// newTestHandler builds the full production HTTP stack (server.NewHandler
// plus the same registrars as cmd/calendar/main.go) against the package-level
// db, using the given TimezoneFinder. Exposed separately from newTestServer
// so a spec can drive it directly via httptest.NewRecorder (no real socket)
// when it needs to run inside a synctest bubble - see auth_test.go's
// constant-time login delay specs.
func newTestHandler(finder handler.TimezoneFinder) http.Handler {
	store, err := bunstore.NewWithCleanupInterval(db, 0)
	Expect(err).NotTo(HaveOccurred())

	sm := server.NewSessionManager(store)

	return server.NewHandler(
		calendar.RegisterAssetsHandler,
		func(mux *http.ServeMux) { handler.NewSetupHandler(db, sm).Register(mux) },
		func(mux *http.ServeMux) { handler.NewSettingsHandler(db, sm).Register(mux) },
		func(mux *http.ServeMux) { handler.NewAuthenticationHandler(db, sm).Register(mux) },
		func(mux *http.ServeMux) { handler.NewEventsHandler(db, sm).Register(mux) },
		func(mux *http.ServeMux) { handler.NewEditEventHandler(db, sm, finder).Register(mux) },
		func(mux *http.ServeMux) { handler.NewUsersHandler(db, sm).Register(mux) },
		func(mux *http.ServeMux) { handler.NewFeedHandler(db).Register(mux) },
	)
}

// newTestServer wraps newTestHandler in a real TLS listener. The session
// cookie is marked Secure, so the server must be TLS and driven via its own
// trusted client - a plain httptest.Server + http.Client would silently
// never send the cookie back.
func newTestServer(finder handler.TimezoneFinder) *httptest.Server {
	return httptest.NewTLSServer(newTestHandler(finder))
}

// newTestClient returns a client trusting ts's certificate, with its own
// fresh cookie jar - simulating one browser/user. Call it once per simulated
// user against the same server.
func newTestClient(ts *httptest.Server) *http.Client {
	jar, err := cookiejar.New(nil)
	Expect(err).NotTo(HaveOccurred())

	client := *ts.Client()
	client.Jar = jar

	return &client
}

// noRedirectClient returns a shallow copy of client (sharing its jar and
// transport) that stops at the first redirect instead of following it, so
// tests can assert on the redirect response itself.
func noRedirectClient(client *http.Client) *http.Client {
	c := *client
	c.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &c
}

// sessionCookieValue returns the current session_id cookie value stored in
// client's jar for ts, or "" if none is set.
func sessionCookieValue(ts *httptest.Server, client *http.Client) string {
	u, err := url.Parse(ts.URL)
	Expect(err).NotTo(HaveOccurred())

	for _, c := range client.Jar.Cookies(u) {
		if c.Name == "session_id" {
			return c.Value
		}
	}

	return ""
}

// mustCreateSettings inserts default settings and returns them.
func mustCreateSettings(ctx context.Context) *domain.Settings {
	settings := domain.NewDefaultSettings()
	Expect(model.InsertSettings(ctx, db, settings)).To(Succeed())
	return settings
}

// mustCreateUser hashes password and inserts a user with the given username
// and role, returning the created domain user (with its plaintext password
// not retained - callers must remember the plaintext themselves for login).
func mustCreateUser(ctx context.Context, username, password string, role domain.Role) *domain.User {
	user := &domain.User{
		ID:       snowflake.Generate(),
		Username: username,
		Role:     role,
	}

	Expect(user.SetPassword(password)).To(Succeed())
	Expect(model.InsertUser(ctx, db, user)).To(Succeed())

	return user
}

// doForm issues a POST with an application/x-www-form-urlencoded body,
// optionally marked as an HTMX request, returning the response (with its
// body drained and closed) and the body content.
func doForm(client *http.Client, ts *httptest.Server, path string, form url.Values, hx bool) (*http.Response, string) {
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, strings.NewReader(form.Encode()))
	Expect(err).NotTo(HaveOccurred())

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if hx {
		req.Header.Set("HX-Request", "true")
	}

	return doRequest(client, req)
}

// doGet issues a GET request with optional query params, optionally marked
// as an HTMX request.
func doGet(client *http.Client, ts *httptest.Server, path string, query url.Values, hx bool) (*http.Response, string) {
	u := ts.URL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u, nil)
	Expect(err).NotTo(HaveOccurred())

	if hx {
		req.Header.Set("HX-Request", "true")
	}

	return doRequest(client, req)
}

func doRequest(client *http.Client, req *http.Request) (*http.Response, string) {
	resp, err := client.Do(req)
	Expect(err).NotTo(HaveOccurred())

	body, err := io.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.Body.Close()).To(Succeed())

	return resp, string(body)
}

// loginAs logs in as username/password using client against ts, asserting
// the login redirects to "/" and establishes a session cookie, then returns
// that cookie's value.
func loginAs(ts *httptest.Server, client *http.Client, username, password string) string {
	GinkgoHelper()

	resp, _ := doForm(noRedirectClient(client), ts, "/login", url.Values{
		"username": {username},
		"password": {password},
	}, false)

	Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
	Expect(resp.Header.Get("Location")).To(Equal("/"))

	cookie := sessionCookieValue(ts, client)
	Expect(cookie).NotTo(BeEmpty())

	return cookie
}
