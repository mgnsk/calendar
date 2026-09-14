package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"time"

	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("events listing", func() {
	var (
		ts      *httptest.Server
		ownerA  *domain.User
		ownerB  *domain.User
		evPast  *domain.Event
		evFutA  *domain.Event
		evDraft *domain.Event
		evFutB  *domain.Event
	)

	BeforeEach(func(ctx SpecContext) {
		mustCreateSettings(ctx)
		ownerA = mustCreateUser(ctx, "ownerA", "secret123", domain.Admin)
		ownerB = mustCreateUser(ctx, "ownerB", "secret123", domain.Author)

		evPast = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now().Add(-2 * time.Hour),
			Title:       "Past Event",
			Description: "A past thing",
			Location:    "Somewhere",
			UserID:      ownerA.ID,
		}
		Expect(model.InsertEvent(ctx, db, evPast)).To(Succeed())

		evFutA = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now().Add(2 * time.Hour),
			Title:       "Future Event A",
			Description: "Owned by A",
			Location:    "Somewhere",
			UserID:      ownerA.ID,
		}
		Expect(model.InsertEvent(ctx, db, evFutA)).To(Succeed())

		evDraft = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now().Add(3 * time.Hour),
			Title:       "Draft Event A",
			Description: "Draft by A",
			Location:    "Somewhere",
			IsDraft:     true,
			UserID:      ownerA.ID,
		}
		Expect(model.InsertEvent(ctx, db, evDraft)).To(Succeed())

		evFutB = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now().Add(4 * time.Hour),
			Title:       "Future Event B",
			Description: "Owned by B",
			Location:    "Somewhere",
			UserID:      ownerB.ID,
		}
		Expect(model.InsertEvent(ctx, db, evFutB)).To(Succeed())

		ts = newTestServer(fakeTimezoneFinder{})
		DeferCleanup(ts.Close)
	})

	Describe("GET /", func() {
		It("shows only future published events, ascending by start time, anonymous", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Future Event A"))
			Expect(body).To(ContainSubstring("Future Event B"))
			Expect(body).NotTo(ContainSubstring("Past Event"))
			Expect(body).NotTo(ContainSubstring("Draft Event A"))

			// Ascending: "Future Event A" (t+2h) appears before "Future Event B" (t+4h).
			Expect(indexOfSubstr(body, "Future Event A")).To(BeNumerically("<", indexOfSubstr(body, "Future Event B")))
		})

		It("returns an HTMX partial without the CSP meta tag or nav", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/", nil, true)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).NotTo(ContainSubstring("Content-Security-Policy"))
			Expect(body).To(ContainSubstring("Future Event A"))
		})

		It("returns the full page with an alphabetically sorted tag cloud for non-HTMX requests", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Content-Security-Policy"))
		})

		It("filters by search term and never returns drafts", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/", url.Values{"search": {"Future"}}, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Future Event A"))
			Expect(body).To(ContainSubstring("Future Event B"))
			Expect(body).NotTo(ContainSubstring("Past Event"))

			respDraft, bodyDraft := doGet(client, ts, "/", url.Values{"search": {"Draft"}}, false)
			Expect(respDraft.StatusCode).To(Equal(http.StatusOK))
			Expect(bodyDraft).NotTo(ContainSubstring("Draft Event A"))
		})

		It("paginates via offset", func(ctx SpecContext) {
			// Seed more than one page worth of future events.
			for i := 0; i < contract.EventLimitPerPage+2; i++ {
				Expect(model.InsertEvent(ctx, db, &domain.Event{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(time.Duration(10+i) * time.Hour),
					Title:       "Bulk Event " + strconv.Itoa(i),
					Description: "bulk",
					Location:    "Somewhere",
					UserID:      ownerA.ID,
				})).To(Succeed())
			}

			client := newTestClient(ts)

			firstPage := doGetBody(client, ts, "/", url.Values{"offset": {"0"}}, true)
			secondPage := doGetBody(client, ts, "/", url.Values{"offset": {"25"}}, true)

			Expect(firstPage).NotTo(Equal(secondPage))
		})
	})

	Describe("GET /past", func() {
		It("shows only past published events, descending by start time", func() {
			client := newTestClient(ts)

			resp, body := doGet(client, ts, "/past", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Past Event"))
			Expect(body).NotTo(ContainSubstring("Future Event A"))
			Expect(body).NotTo(ContainSubstring("Draft Event A"))
		})
	})

	Describe("GET /my-events", func() {
		It("is forbidden for anonymous users", func() {
			client := newTestClient(ts)

			resp, _ := doGet(client, ts, "/my-events", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusForbidden))
		})

		It("shows the owner's own events including their own drafts, excluding other users' events", func() {
			client := newTestClient(ts)
			loginAs(ts, client, "ownerA", "secret123")

			resp, body := doGet(client, ts, "/my-events", nil, false)

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("Future Event A"))
			Expect(body).To(ContainSubstring("Draft Event A"))
			Expect(body).To(ContainSubstring("Past Event"))
			Expect(body).NotTo(ContainSubstring("Future Event B"))
		})

		It("paginates via last_id, in created-at descending order", func(ctx SpecContext) {
			bulkIDs := make([]snowflake.ID, contract.EventLimitPerPage+2)

			for i := range bulkIDs {
				bulkIDs[i] = snowflake.Generate()

				Expect(model.InsertEvent(ctx, db, &domain.Event{
					ID:          bulkIDs[i],
					StartAt:     time.Now().Add(time.Duration(10+i) * time.Hour),
					Title:       "My Bulk Event " + strconv.Itoa(i),
					Description: "bulk",
					Location:    "Somewhere",
					UserID:      ownerA.ID,
				})).To(Succeed())
			}

			client := newTestClient(ts)
			loginAs(ts, client, "ownerA", "secret123")

			// Newest-created first: page 1 (no cursor) holds the 25 most
			// recently created bulk events (indices 26 down to 2).
			firstPage := doGetBody(client, ts, "/my-events", nil, true)
			Expect(firstPage).To(ContainSubstring("My Bulk Event 26"))
			Expect(firstPage).To(ContainSubstring("My Bulk Event 2"))
			Expect(firstPage).NotTo(ContainSubstring("My Bulk Event 1<"))
			Expect(firstPage).NotTo(ContainSubstring("Past Event"))

			// Page 2, cursored past the last item of page 1: the two oldest
			// bulk events plus the pre-existing fixtures for ownerA.
			secondPage := doGetBody(client, ts, "/my-events", url.Values{
				"last_id": {strconv.FormatInt(bulkIDs[2].Int64(), 10)},
			}, true)
			Expect(secondPage).To(ContainSubstring("My Bulk Event 1<"))
			Expect(secondPage).To(ContainSubstring("My Bulk Event 0"))
			Expect(secondPage).To(ContainSubstring("Draft Event A"))
			Expect(secondPage).To(ContainSubstring("Past Event"))
			Expect(secondPage).NotTo(ContainSubstring("My Bulk Event 2<"))
		})
	})
})

func indexOfSubstr(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func doGetBody(client *http.Client, ts *httptest.Server, path string, query url.Values, hx bool) string {
	_, body := doGet(client, ts, path, query, hx)
	return body
}
