package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/api"
	"github.com/mgnsk/calendar/internal/model"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	"github.com/mmcdole/gofeed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("RSS feed output", func() {
	var (
		rec *httptest.ResponseRecorder
		c   echo.Context
		h   *api.FeedHandler
	)

	BeforeEach(func(ctx SpecContext) {
		By("inserting events", func() {
			Expect(model.InsertEvent(ctx, db, event1)).To(Succeed())
			Expect(model.InsertEvent(ctx, db, event2)).To(Succeed())
			Expect(model.InsertEvent(ctx, db, event3)).To(Succeed())
		})

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		h = api.NewFeedHandler(db, api.FeedConfig{
			Title: "My Test Feed",
			Link:  "https://example.testing",
		})
	})

	DescribeTable("feed types",
		func(handle echo.HandlerFunc, feedType api.FeedType, contentType string) {
			Expect(handle(c)).To(Succeed())

			r := rec.Result()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			Expect(r.Header).To(SatisfyAll(
				HaveKeyWithValue(echo.HeaderContentType, HaveExactElements(
					Equal(contentType),
				)),
				HaveKeyWithValue(echo.HeaderContentDisposition, HaveExactElements(
					Equal(`attachment; filename="feed.xml"`),
				)),
			))

			fp := gofeed.NewParser()
			feed := Must(fp.Parse(r.Body))

			Expect(feed).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"FeedType": Equal(string(feedType)),
				"Title":    Equal("My Test Feed"),
				"Link":     Equal("https://example.testing"),
				"Items": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Title":           Equal(event3.Title),
						"Description":     Equal(event3.GetDescription()),
						"PublishedParsed": PointTo(BeTemporally("~", event3.StartAt.Time(), time.Second)),
						"GUID":            Equal(event3.ID.String()),
						"Link":            Equal(event3.URL),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Title":           Equal(event2.Title),
						"Description":     Equal(event2.GetDescription()),
						"PublishedParsed": PointTo(BeTemporally("~", event2.StartAt.Time(), time.Second)),
						"GUID":            Equal(event2.ID.String()),
						"Link":            Equal(event2.URL),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Title":           Equal(event1.Title),
						"Description":     Equal(event1.GetDescription()),
						"PublishedParsed": PointTo(BeTemporally("~", event1.StartAt.Time(), time.Second)),
						"GUID":            Equal(event1.ID.String()),
						"Link":            Equal(event1.URL),
					})),
				),
			})))
		},

		Entry(
			"RSS feed",
			func(c echo.Context) error {
				return h.CreateFeedHandler(api.RSS)(c)
			},
			api.RSS,
			"application/rss+xml; charset=utf-8",
		),

		Entry("Atom feed",
			func(c echo.Context) error {
				return h.CreateFeedHandler(api.Atom)(c)
			},
			api.Atom,
			"application/atom+xml; charset=utf-8",
		),
	)
})
