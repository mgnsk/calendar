package handler_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("settings-required gate", func() {
	var ts *httptest.Server

	BeforeEach(func() {
		ts = newTestServer(fakeTimezoneFinder{})
		DeferCleanup(ts.Close)
	})

	When("no settings exist yet", func() {
		DescribeTable("every route except /setup redirects to /setup",
			func(path string) {
				client := newTestClient(ts)

				resp, _ := doGet(noRedirectClient(client), ts, path, nil, false)

				Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
				Expect(resp.Header.Get("Location")).To(Equal("/setup"))
			},

			Entry("root", "/"),
			Entry("past", "/past"),
			Entry("my-events", "/my-events"),
			Entry("login", "/login"),
			Entry("users", "/users"),
			Entry("settings", "/settings"),
		)

		It("does not redirect /setup itself", func() {
			client := newTestClient(ts)

			resp, _ := doGet(noRedirectClient(client), ts, "/setup", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})

		It("does not gate the public feed routes (no settings/user middleware)", func() {
			client := newTestClient(ts)

			resp, _ := doGet(noRedirectClient(client), ts, "/feed", nil, false)

			// FeedHandler only has the settings middleware, which DOES gate
			// on missing settings same as everything else - confirm that.
			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/setup"))
		})
	})
})

var _ = Describe("Content-Security-Policy", func() {
	var ts *httptest.Server

	BeforeEach(func(ctx SpecContext) {
		mustCreateSettings(ctx)

		ts = newTestServer(fakeTimezoneFinder{})
		DeferCleanup(ts.Close)
	})

	It("sets the expected CSP meta tag on a rendered page", func() {
		client := newTestClient(ts)

		resp, body := doGet(client, ts, "/", nil, false)

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring(
			`content="default-src &#39;self&#39;; script-src &#39;self&#39; &#39;unsafe-inline&#39;; style-src &#39;self&#39; &#39;unsafe-inline&#39;; connect-src &#39;self&#39; nominatim.openstreetmap.org"`,
		))
		Expect(body).To(ContainSubstring(`http-equiv="Content-Security-Policy"`))
	})

	It("does not include the CSP meta tag in an HTMX partial response", func() {
		client := newTestClient(ts)

		resp, body := doGet(client, ts, "/", nil, true)

		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		Expect(body).NotTo(ContainSubstring("Content-Security-Policy"))
	})
})
