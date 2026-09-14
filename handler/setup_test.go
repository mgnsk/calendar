package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/model"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("setup", func() {
	var ts *httptest.Server

	BeforeEach(func() {
		ts = newTestServer(fakeTimezoneFinder{})
		DeferCleanup(ts.Close)
	})

	When("no settings exist yet", func() {
		It("renders the setup form on GET", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/setup", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring(`name="password1"`))
		})

		DescribeTable("POST validation errors",
			func(ctx SpecContext, form url.Values, wantErrorMessage string) {
				client := newTestClient(ts)

				resp, body := doForm(client, ts, "/setup", form, false)

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(body).To(ContainSubstring(wantErrorMessage))

				_, err := model.GetSettings(ctx, db)
				Expect(err).To(MatchError(calendar.NotFound))
			},

			Entry("missing title", url.Values{
				"pagetitle": {""},
				"username":  {"admin"},
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, "Title must be set"),

			Entry("missing username", url.Values{
				"pagetitle": {"My Calendar"},
				"username":  {""},
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, "Username must be set"),

			Entry("mismatched passwords", url.Values{
				"pagetitle": {"My Calendar"},
				"username":  {"admin"},
				"password1": {"secret123"},
				"password2": {"different"},
			}, "Passwords must match"),
		)

		It("creates settings and the first admin user, then logs them in", func(ctx SpecContext) {
			client := newTestClient(ts)

			resp, _ := doForm(noRedirectClient(client), ts, "/setup", url.Values{
				"pagetitle": {"My Calendar"},
				"pagedesc":  {"My Calendar Description"},
				"username":  {"admin"},
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
			Expect(sessionCookieValue(ts, client)).NotTo(BeEmpty())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings.Title).To(Equal("My Calendar"))
			Expect(settings.Description).To(Equal("My Calendar Description"))

			user := Must(model.GetUserByUsername(ctx, db, "admin"))
			Expect(user.Role).To(BeEquivalentTo("admin"))

			// The new admin can now reach an admin-only page.
			settingsResp, _ := doGet(client, ts, "/settings", nil, false)
			Expect(settingsResp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	When("settings already exist", func() {
		BeforeEach(func(ctx SpecContext) {
			mustCreateSettings(ctx)
		})

		It("redirects GET /setup to /", func() {
			client := newTestClient(ts)

			resp, _ := doGet(noRedirectClient(client), ts, "/setup", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
		})

		It("redirects POST /setup to / without creating a second admin", func(ctx SpecContext) {
			client := newTestClient(ts)

			resp, _ := doForm(noRedirectClient(client), ts, "/setup", url.Values{
				"pagetitle": {"Ignored"},
				"username":  {"someone"},
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))

			_, err := model.GetUserByUsername(ctx, db, "someone")
			Expect(err).To(MatchError(calendar.NotFound))
		})
	})
})
