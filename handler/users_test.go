package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func extractInviteToken(body string) string {
	const marker = `href="/register/`
	i := strings.Index(body, marker)
	if i < 0 {
		return ""
	}
	rest := body[i+len(marker):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

var _ = Describe("users management", func() {
	var (
		ts      *httptest.Server
		admin   *domain.User
		author1 *domain.User
	)

	BeforeEach(func(ctx SpecContext) {
		mustCreateSettings(ctx)
		admin = mustCreateUser(ctx, "admin", "secret123", domain.Admin)
		author1 = mustCreateUser(ctx, "author1", "secret123", domain.Author)

		ts = newTestServer(fakeTimezoneFinder{})
		DeferCleanup(ts.Close)
	})

	Describe("GET /users", func() {
		It("is forbidden for anonymous users", func() {
			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/users", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for non-admin users", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doGet(client, ts, "/users", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("lists all users for an admin", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, body := doGet(client, ts, "/users", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("admin"))
			Expect(body).To(ContainSubstring("author1"))
		})
	})

	Describe("POST /invite", func() {
		It("is forbidden for non-admin users", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(client, ts, "/invite", url.Values{}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("returns 404 for admin when not an HTMX request", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/invite", url.Values{}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("creates a valid invite for an admin via HTMX", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, body := doForm(client, ts, "/invite", url.Values{}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			token := extractInviteToken(body)
			Expect(token).NotTo(BeEmpty())

			invite := Must(model.GetInvite(ctx, db, uuid.MustParse(token)))
			Expect(invite.IsValid()).To(BeTrue())
			Expect(time.Until(invite.ValidUntil)).To(BeNumerically("~", 72*time.Hour, time.Minute))
		})
	})

	Describe("GET/POST /register/{token}", func() {
		It("redirects to / when already authenticated", func(ctx SpecContext) {
			token := uuid.New()
			Expect(model.InsertInvite(ctx, db, &domain.Invite{
				Token:      token,
				ValidUntil: time.Now().Add(time.Hour),
				CreatedBy:  admin.ID,
			})).To(Succeed())

			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doGet(noRedirectClient(client), ts, "/register/"+token.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
		})

		It("returns 404 for a well-formed but unknown token", func() {
			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/register/"+uuid.New().String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("returns 400 for a malformed token", func() {
			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/register/not-a-uuid", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("returns 404 for an expired invite", func(ctx SpecContext) {
			token := uuid.New()
			Expect(model.InsertInvite(ctx, db, &domain.Invite{
				Token:      token,
				ValidUntil: time.Now().Add(-time.Hour),
				CreatedBy:  admin.ID,
			})).To(Succeed())

			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/register/"+token.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("renders the registration form for a valid invite", func(ctx SpecContext) {
			token := uuid.New()
			Expect(model.InsertInvite(ctx, db, &domain.Invite{
				Token:      token,
				ValidUntil: time.Now().Add(time.Hour),
				CreatedBy:  admin.ID,
			})).To(Succeed())

			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/register/"+token.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring(`name="password1"`))
		})

		DescribeTable("POST validation errors",
			func(ctx SpecContext, form url.Values, wantErrorMessage string) {
				token := uuid.New()
				Expect(model.InsertInvite(ctx, db, &domain.Invite{
					Token:      token,
					ValidUntil: time.Now().Add(time.Hour),
					CreatedBy:  admin.ID,
				})).To(Succeed())

				client := newTestClient(ts)

				resp, body := doForm(client, ts, "/register/"+token.String(), form, false)

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(body).To(ContainSubstring(wantErrorMessage))
			},

			Entry("missing username", url.Values{
				"username":  {""},
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, "Username must be set"),

			Entry("mismatched passwords", url.Values{
				"username":  {"newuser"},
				"password1": {"secret123"},
				"password2": {"different"},
			}, "Passwords must match"),
		)

		It("re-renders with an error and keeps the invite valid on duplicate username", func(ctx SpecContext) {
			token := uuid.New()
			Expect(model.InsertInvite(ctx, db, &domain.Invite{
				Token:      token,
				ValidUntil: time.Now().Add(time.Hour),
				CreatedBy:  admin.ID,
			})).To(Succeed())

			client := newTestClient(ts)

			resp, body := doForm(client, ts, "/register/"+token.String(), url.Values{
				"username":  {"author1"}, // already exists
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("User already exists"))

			// Invite must still be usable (the failed tx must have rolled back).
			invite := Must(model.GetInvite(ctx, db, token))
			Expect(invite.IsValid()).To(BeTrue())
		})

		It("registers a new author user and consumes the invite on success", func(ctx SpecContext) {
			token := uuid.New()
			Expect(model.InsertInvite(ctx, db, &domain.Invite{
				Token:      token,
				ValidUntil: time.Now().Add(time.Hour),
				CreatedBy:  admin.ID,
			})).To(Succeed())

			client := newTestClient(ts)

			resp, _ := doForm(noRedirectClient(client), ts, "/register/"+token.String(), url.Values{
				"username":  {"newuser"},
				"password1": {"secret123"},
				"password2": {"secret123"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/"))
			Expect(sessionCookieValue(ts, client)).NotTo(BeEmpty())

			newUser := Must(model.GetUserByUsername(ctx, db, "newuser"))
			Expect(newUser.Role).To(Equal(domain.Author))

			_, err := model.GetInvite(ctx, db, token)
			Expect(err).To(MatchError(calendar.NotFound))
		})
	})

	Describe("POST /delete-user", func() {
		It("is forbidden for anonymous users", func() {
			client := newTestClient(ts)

			resp, _ := doForm(client, ts, "/delete-user", url.Values{
				"user_id": {author1.ID.String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for non-admin users", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(client, ts, "/delete-user", url.Values{
				"user_id": {admin.ID.String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for an admin targeting themselves, even without HTMX", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/delete-user", url.Values{
				"user_id": {admin.ID.String()},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("returns 404 for admin targeting another user when not HTMX", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/delete-user", url.Values{
				"user_id": {author1.ID.String()},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("deletes the target user for admin via HTMX", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/delete-user", url.Values{
				"user_id": {author1.ID.String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("HX-Refresh")).To(Equal("true"))

			_, err := model.GetUserByUsername(ctx, db, "author1")
			Expect(err).To(MatchError(calendar.NotFound))
		})

		It("returns 404 for a nonexistent target user id", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/delete-user", url.Values{
				"user_id": {snowflake.Generate().String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("POST /upgrade-user", func() {
		It("is forbidden for anonymous users", func() {
			client := newTestClient(ts)

			resp, _ := doForm(client, ts, "/upgrade-user", url.Values{
				"user_id": {author1.ID.String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for non-admin users", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(client, ts, "/upgrade-user", url.Values{
				"user_id": {author1.ID.String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for an admin targeting themselves, even without HTMX", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/upgrade-user", url.Values{
				"user_id": {admin.ID.String()},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("returns 404 for admin targeting another user when not HTMX", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/upgrade-user", url.Values{
				"user_id": {author1.ID.String()},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("promotes the target user to admin via HTMX", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/upgrade-user", url.Values{
				"user_id": {author1.ID.String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("HX-Refresh")).To(Equal("true"))

			updated := Must(model.GetUser(ctx, db, author1.ID))
			Expect(updated.Role).To(Equal(domain.Admin))

			// The newly promoted admin can now reach an admin-only page.
			author1Client := newTestClient(ts)
			loginAs(ts, author1Client, "author1", "secret123")
			settingsResp, _ := doGet(author1Client, ts, "/settings", nil, false)
			Expect(settingsResp.StatusCode).To(Equal(http.StatusOK))
		})

		It("returns 404 for a nonexistent target user id (GetUser fails before UpdateUser)", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(client, ts, "/upgrade-user", url.Values{
				"user_id": {snowflake.Generate().String()},
			}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})
})
