package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("edit event", func() {
	var (
		ts       *httptest.Server
		author1  *domain.User
		existing *domain.Event
	)

	BeforeEach(func(ctx SpecContext) {
		mustCreateSettings(ctx)
		author1 = mustCreateUser(ctx, "author1", "secret123", domain.Author)
		mustCreateUser(ctx, "author2", "secret123", domain.Author)
		mustCreateUser(ctx, "admin", "secret123", domain.Admin)

		existing = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now().Add(2 * time.Hour),
			Title:       "Existing Event",
			Description: "Existing description",
			Location:    "Tallinn",
			UserID:      author1.ID,
		}
		Expect(model.InsertEvent(ctx, db, existing)).To(Succeed())

		ts = newTestServer(fakeTimezoneFinder{})
		DeferCleanup(ts.Close)
	})

	Describe("GET /edit/{event_id}", func() {
		It("is forbidden for anonymous users creating a new event", func() {
			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/edit/0", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("renders an empty form for a logged-in user creating a new event", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, body := doGet(client, ts, "/edit/0", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring(`name="title"`))
		})

		It("is forbidden for anonymous users editing an existing event", func() {
			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/edit/"+existing.ID.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for a non-admin, non-owner editing someone else's event", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author2", "secret123")

			resp, _ := doGet(client, ts, "/edit/"+existing.ID.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("allows an admin to edit someone else's event", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, body := doGet(client, ts, "/edit/"+existing.ID.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Existing Event"))
		})

		It("prefills the form for the owner", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, body := doGet(client, ts, "/edit/"+existing.ID.String(), nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Existing Event"))
			Expect(body).To(ContainSubstring("Existing description"))
		})

		It("returns 404 for a nonexistent numeric event id", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doGet(client, ts, "/edit/999999999999999999", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("returns 400 for a malformed event id", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doGet(client, ts, "/edit/not-a-number", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("POST /edit/0 (create)", func() {
		DescribeTable("validation errors re-render the form",
			func(form url.Values, wantErrorMessage string) {
				client := newTestClient(ts)
				loginAs(ts, client, "author1", "secret123")

				resp, body := doForm(client, ts, "/edit/0", form, false)

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(body).To(ContainSubstring(wantErrorMessage))
			},

			Entry("missing title", url.Values{
				"desc":     {"Description"},
				"start_at": {"2030-06-01T10:00"},
				"location": {"Tallinn"},
			}, "Required"),

			Entry("missing description", url.Values{
				"title":    {"Title"},
				"start_at": {"2030-06-01T10:00"},
				"location": {"Tallinn"},
			}, "Required"),

			Entry("missing start_at", url.Values{
				"title":    {"Title"},
				"desc":     {"Description"},
				"location": {"Tallinn"},
			}, "Required"),

			Entry("missing location", url.Values{
				"title":    {"Title"},
				"desc":     {"Description"},
				"start_at": {"2030-06-01T10:00"},
			}, "Required"),

			Entry("malformed start_at", url.Values{
				"title":    {"Title"},
				"desc":     {"Description"},
				"start_at": {"not-a-date"},
				"location": {"Tallinn"},
			}, "Invalid format"),

			Entry("invalid URL", url.Values{
				"title":    {"Title"},
				"desc":     {"Description"},
				"start_at": {"2030-06-01T10:00"},
				"location": {"Tallinn"},
				"url":      {"http://example.com/\n"},
			}, "Invalid URL"),
		)

		It("creates a published event visible on / and redirects to /edit/{id}", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(noRedirectClient(client), ts, "/edit/0", url.Values{
				"title":    {"New Published Event"},
				"desc":     {"A new event"},
				"start_at": {"2030-06-01T10:00"},
				"location": {"Tallinn"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			location := resp.Header.Get("Location")
			Expect(location).To(HavePrefix("/edit/"))

			listResp, listBody := doGet(newTestClient(ts), ts, "/", nil, false)
			Expect(listResp.StatusCode).To(Equal(http.StatusOK))
			Expect(listBody).To(ContainSubstring("New Published Event"))
		})

		It("creates a draft visible only on /my-events for the owner", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(noRedirectClient(client), ts, "/edit/0?draft=true", url.Values{
				"title":    {"New Draft Event"},
				"desc":     {"A draft event"},
				"start_at": {"2030-06-01T10:00"},
				"location": {"Tallinn"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))

			anonResp, anonBody := doGet(newTestClient(ts), ts, "/", nil, false)
			Expect(anonResp.StatusCode).To(Equal(http.StatusOK))
			Expect(anonBody).NotTo(ContainSubstring("New Draft Event"))

			myEventsResp, myEventsBody := doGet(client, ts, "/my-events", nil, false)
			Expect(myEventsResp.StatusCode).To(Equal(http.StatusOK))
			Expect(myEventsBody).To(ContainSubstring("New Draft Event"))
		})
	})

	Describe("POST /edit/{event_id} (update)", func() {
		It("lets the owner update their own event", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(noRedirectClient(client), ts, "/edit/"+existing.ID.String(), url.Values{
				"title":    {"Updated Title"},
				"desc":     {"Updated description"},
				"start_at": {"2030-07-01T11:00"},
				"location": {"Tallinn"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))
			Expect(resp.Header.Get("Location")).To(Equal("/edit/" + existing.ID.String()))

			updated := Must(model.GetEvent(ctx, db, existing.ID))
			Expect(updated.Title).To(Equal("Updated Title"))
			Expect(updated.UserID).To(Equal(author1.ID))
		})

		It("lets an admin update someone else's event without changing ownership", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "admin", "secret123")

			resp, _ := doForm(noRedirectClient(client), ts, "/edit/"+existing.ID.String(), url.Values{
				"title":    {"Admin Edited Title"},
				"desc":     {"Updated by admin"},
				"start_at": {"2030-07-01T11:00"},
				"location": {"Tallinn"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))

			updated := Must(model.GetEvent(ctx, db, existing.ID))
			Expect(updated.Title).To(Equal("Admin Edited Title"))
			Expect(updated.UserID).To(Equal(author1.ID))
		})
	})

	Describe("timezone resolution", func() {
		It("uses the geo-resolved timezone over the user's browser timezone", func(ctx SpecContext) {
			tzTS := newTestServer(fakeTimezoneFinder{name: "America/New_York"})
			DeferCleanup(tzTS.Close)

			client := newTestClient(tzTS)
			loginAs(tzTS, client, "author1", "secret123")

			resp, _ := doForm(noRedirectClient(client), tzTS, "/edit/0", url.Values{
				"title":         {"TZ Event"},
				"desc":          {"Desc"},
				"start_at":      {"2030-06-01T10:00"},
				"location":      {"NYC"},
				"user_timezone": {"Europe/Tallinn"}, // must be ignored: geo finder wins
			}, false)
			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))

			loc, err := time.LoadLocation("America/New_York")
			Expect(err).NotTo(HaveOccurred())
			expected, err := time.ParseInLocation("2006-01-02T15:04", "2030-06-01T10:00", loc)
			Expect(err).NotTo(HaveOccurred())

			id := lastPathSegment(resp.Header.Get("Location"))
			ev := Must(model.GetEvent(ctx, db, mustParseSnowflakeID(id)))
			Expect(ev.StartAt.Unix()).To(Equal(expected.Unix()))
		})

		It("falls back to the browser's user_timezone when geo lookup is empty", func(ctx SpecContext) {
			client := newTestClient(ts) // fakeTimezoneFinder{} always returns ""
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(noRedirectClient(client), ts, "/edit/0", url.Values{
				"title":         {"TZ Fallback Event"},
				"desc":          {"Desc"},
				"start_at":      {"2030-06-01T10:00"},
				"location":      {"Somewhere"},
				"user_timezone": {"America/New_York"},
			}, false)
			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))

			loc, err := time.LoadLocation("America/New_York")
			Expect(err).NotTo(HaveOccurred())
			expected, err := time.ParseInLocation("2006-01-02T15:04", "2030-06-01T10:00", loc)
			Expect(err).NotTo(HaveOccurred())

			id := lastPathSegment(resp.Header.Get("Location"))
			ev := Must(model.GetEvent(ctx, db, mustParseSnowflakeID(id)))
			Expect(ev.StartAt.Unix()).To(Equal(expected.Unix()))
		})

		It("falls back to UTC when neither geo lookup nor user_timezone are set", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(noRedirectClient(client), ts, "/edit/0", url.Values{
				"title":    {"UTC Fallback Event"},
				"desc":     {"Desc"},
				"start_at": {"2030-06-01T10:00"},
				"location": {"Somewhere"},
			}, false)
			Expect(resp.StatusCode).To(Equal(http.StatusSeeOther))

			expected, err := time.ParseInLocation("2006-01-02T15:04", "2030-06-01T10:00", time.UTC)
			Expect(err).NotTo(HaveOccurred())

			id := lastPathSegment(resp.Header.Get("Location"))
			ev := Must(model.GetEvent(ctx, db, mustParseSnowflakeID(id)))
			Expect(ev.StartAt.Unix()).To(Equal(expected.Unix()))
		})

		It("re-renders the form with an error for an invalid user_timezone", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, body := doForm(client, ts, "/edit/0", url.Values{
				"title":         {"Bad TZ Event"},
				"desc":          {"Desc"},
				"start_at":      {"2030-06-01T10:00"},
				"location":      {"Somewhere"},
				"user_timezone": {"Not/AZone"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Invalid start_at value"))
		})
	})

	Describe("POST /delete/{event_id}", func() {
		It("is forbidden for anonymous users", func() {
			client := newTestClient(ts)

			resp, _ := doForm(client, ts, "/delete/"+existing.ID.String(), url.Values{}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("is forbidden for a non-owner, non-admin", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author2", "secret123")

			resp, _ := doForm(client, ts, "/delete/"+existing.ID.String(), url.Values{}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("returns 404 for the owner when not an HTMX request", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(client, ts, "/delete/"+existing.ID.String(), url.Values{}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("deletes the event for the owner via HTMX and sets HX-Refresh", func(ctx SpecContext) {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(client, ts, "/delete/"+existing.ID.String(), url.Values{}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("HX-Refresh")).To(Equal("true"))

			_, err := model.GetEvent(ctx, db, existing.ID)
			Expect(err).To(MatchError(calendar.NotFound))
		})

		It("returns 404 for a nonexistent event id", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, _ := doForm(client, ts, "/delete/999999999999999999", url.Values{}, true)

			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})
	})

	Describe("POST /preview", func() {
		It("is forbidden for anonymous users", func() {
			client := newTestClient(ts)

			resp, _ := doForm(client, ts, "/preview", url.Values{}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("renders a card fragment without panicking, even with an invalid start_at", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "author1", "secret123")

			resp, body := doForm(client, ts, "/preview", url.Values{
				"title":    {"Preview Title"},
				"desc":     {"Preview description"},
				"start_at": {"garbage"},
				"location": {"Somewhere"},
			}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Preview Title"))
		})
	})
})

func lastPathSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

func mustParseSnowflakeID(s string) snowflake.ID {
	var id snowflake.ID
	Expect(id.UnmarshalText([]byte(s))).To(Succeed())
	return id
}
