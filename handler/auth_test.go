package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/synctest"

	"github.com/mgnsk/calendar/domain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("authentication", func() {
	var (
		ts *httptest.Server
		h  http.Handler
	)

	BeforeEach(func(ctx SpecContext) {
		mustCreateSettings(ctx)
		mustCreateUser(ctx, "alice", "secret123", domain.Author)

		// Built here, outside any synctest bubble: server.NewSessionManager
		// (via scs.New) always spawns a throwaway memstore cleanup goroutine
		// internally before we replace its Store - see the DescribeTable
		// below. A goroutine spawned outside a bubble is simply not part of
		// it, so it can't trip synctest's leaked-goroutine deadlock check.
		h = newTestHandler(fakeTimezoneFinder{})
		ts = httptest.NewTLSServer(h)
		DeferCleanup(ts.Close)
	})

	Describe("GET /login", func() {
		It("renders the login form when anonymous", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/login", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring(`name="username"`))
		})

		It("redirects to / when already authenticated", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "alice", "secret123")

			resp, _ := doGet(noRedirectClient(client), ts, "/login", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
		})
	})

	Describe("POST /login", func() {
		It("re-renders the form with required-field errors on empty submission", func() {
			client := newTestClient(ts)

			resp, body := doForm(client, ts, "/login", url.Values{
				"username": {""},
				"password": {""},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Required"))
			Expect(sessionCookieValue(ts, client)).To(BeEmpty())
		})

		DescribeTable("invalid credentials return a generic error",
			// The handler waits out a hardcoded ~3s constant-time grace
			// timeout on this path (see handler/auth.go). There's no
			// application-level test seam to shorten it, but testing/synctest
			// virtualizes time for anything driven synchronously in-process:
			// the ~3s wait resolves in a few milliseconds of real wall-clock
			// time. This only works because we drive the handler directly
			// via ServeHTTP (no real socket, using the h built in BeforeEach)
			// instead of going through ts's httptest.NewTLSServer - synctest's
			// fake clock only advances when every goroutine in the bubble is
			// durably blocked, and a goroutine blocked on real network I/O
			// never counts as durably blocked, which deadlocks the bubble
			// instead of fast-forwarding it.
			func(username, password string) {
				synctest.Test(suiteT, func(t *testing.T) {
					req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
						"username": {username},
						"password": {password},
					}.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

					rec := httptest.NewRecorder()
					h.ServeHTTP(rec, req)

					if rec.Code != http.StatusOK {
						t.Fatalf("expected status 200, got %d", rec.Code)
					}
					if !strings.Contains(rec.Body.String(), "Invalid username or password") {
						t.Fatalf("expected body to contain the generic invalid-credentials error, got: %s", rec.Body.String())
					}
					for _, c := range rec.Result().Cookies() {
						if c.Name == "session_id" {
							t.Fatalf("expected no session_id cookie to be set, got one")
						}
					}
				})
			},

			Entry("unknown username", "nobody", "whatever"),
			Entry("wrong password", "alice", "wrongpassword"),
		)

		It("logs in with correct credentials and sets a session cookie", func() {
			client := newTestClient(ts)

			resp, _ := doForm(noRedirectClient(client), ts, "/login", url.Values{
				"username": {"alice"},
				"password": {"secret123"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
			Expect(sessionCookieValue(ts, client)).NotTo(BeEmpty())
		})

		It("renews the session token on each login (session-fixation protection)", func() {
			client := newTestClient(ts)

			firstCookie := loginAs(ts, client, "alice", "secret123")

			doGet(client, ts, "/logout", nil, false)

			secondCookie := loginAs(ts, client, "alice", "secret123")

			Expect(secondCookie).NotTo(Equal(firstCookie))
		})
	})

	Describe("GET /logout", func() {
		It("destroys the session and redirects to /", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "alice", "secret123")

			resp, _ := doGet(noRedirectClient(client), ts, "/logout", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))

			// The user is now anonymous - my-events (login required) is forbidden.
			myEventsResp, _ := doGet(client, ts, "/my-events", nil, false)
			Expect(myEventsResp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("succeeds without a panic when already anonymous", func() {
			client := newTestClient(ts)

			resp, _ := doGet(noRedirectClient(client), ts, "/logout", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
		})
	})
})
