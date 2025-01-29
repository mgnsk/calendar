package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/api"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
	"github.com/mmcdole/gofeed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"golang.org/x/net/context"
)

var _ = Describe("RSS feed output", func() {
	var (
		rec *httptest.ResponseRecorder
		c   echo.Context
		h   *api.FeedHandler
	)

	BeforeEach(func(ctx SpecContext) {
		insertEvents(ctx)

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec = httptest.NewRecorder()
		c = e.NewContext(req, rec)

		h = api.NewFeedHandler(db, api.FeedConfig{
			Title:       "My Test Feed",
			Description: "Feed of awesome events",
			Link:        "https://example.testing",
		})
	})

	DescribeTable("feeds",
		func(handle echo.HandlerFunc, contentType string) {
			Expect(handle(c)).To(Succeed())

			r := rec.Result()
			Expect(r.StatusCode).To(Equal(http.StatusOK))
			Expect(r.Header).To(SatisfyAll(
				HaveKeyWithValue(echo.HeaderContentType, HaveExactElements(
					Equal(contentType),
				)),
			))

			assertFeed(r.Body)
		},
		Entry(
			"RSS feed",
			func(c echo.Context) error {
				return h.CreateFeedHandler(api.RSS)(c)
			},
			"application/rss+xml; charset=utf-8",
		),
		Entry("Atom feed",
			func(c echo.Context) error {
				return h.CreateFeedHandler(api.Atom)(c)
			},
			"application/atom+xml; charset=utf-8",
		),
	)
})

func insertEvents(ctx context.Context) {
	By("inserting events", func() {
		events := []*domain.Event{
			{
				ID:          snowflake.Generate(),
				StartAt:     timestamp.New(baseTime.Add(3 * time.Hour)),
				EndAt:       timestamp.Timestamp{},
				Title:       "Event 1",
				Description: "Desc 1",
				URL:         "",
				Tags:        []string{"tag1"},
			},
			{
				ID:          snowflake.Generate(),
				StartAt:     timestamp.New(baseTime.Add(2 * time.Hour)),
				EndAt:       timestamp.Timestamp{},
				Title:       "Event 2",
				Description: "Desc 2",
				URL:         "",
				Tags:        []string{"tag1", "tag2"},
			},
			{
				ID:          snowflake.Generate(),
				StartAt:     timestamp.New(baseTime.Add(1 * time.Hour)),
				EndAt:       timestamp.Timestamp{},
				Title:       "Event 3",
				Description: "Desc 3",
				URL:         "",
				Tags:        []string{"tag3"},
			},
		}

		for _, ev := range events {
			Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
		}
	})
}

func assertFeed(r io.Reader) {
	fp := gofeed.NewParser()
	feed, err := fp.Parse(r)
	Expect(err).NotTo(HaveOccurred())

	Expect(feed).To(PointTo(MatchFields(IgnoreExtras, Fields{
		"Title":       Equal("My Test Feed"),
		"Description": Equal("Feed of awesome events"),
		"Link":        Equal("https://example.testing"),
		"Items": HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
				"Description": Equal(`Desc 3

tags: tag3
starts at: 2025-01-29T20:55:00+02:00`),
				"PublishedParsed": PointTo(BeTemporally("~", baseTime.Add(time.Hour), time.Second)),
				"GUID":            Not(BeEmpty()),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"Description": Equal(`Desc 2

tags: tag1, tag2
starts at: 2025-01-29T21:55:00+02:00`),
				"PublishedParsed": PointTo(BeTemporally("~", baseTime.Add(2*time.Hour), time.Second)),
				"GUID":            Not(BeEmpty()),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
				"Description": Equal(`Desc 1

tags: tag1
starts at: 2025-01-29T22:55:00+02:00`),
				"PublishedParsed": PointTo(BeTemporally("~", baseTime.Add(3*time.Hour), time.Second)),
				"GUID":            Not(BeEmpty()),
			})),
		),
	})))
}

// baseTime for events.
var baseTime time.Time

func init() {
	loc, err := time.LoadLocation("Europe/Tallinn")
	if err != nil {
		panic(err)
	}
	baseTime = time.Date(2025, 1, 29, 19, 55, 00, 00, loc)
}
