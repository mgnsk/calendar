package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/api"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	"github.com/mmcdole/gofeed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gcustom"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("RSS feed output", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func(ctx SpecContext) {
		By("inserting events", func() {
			Expect(model.InsertEvent(ctx, db, event1)).To(Succeed())
			Expect(model.InsertEvent(ctx, db, event2)).To(Succeed())
			Expect(model.InsertEvent(ctx, db, event3)).To(Succeed())
		})

		e := echo.New()
		h := api.NewFeedHandler(db, api.Config{
			PageTitle: "My Test Feed",
			BaseURL:   Must(url.Parse("https://example.testing")),
		})
		h.Register(e)

		server = httptest.NewServer(e)
		DeferCleanup(server.Close)
	})

	DescribeTable("feed types",
		func(feedType, path, contentType string) {
			r := Must(server.Client().Get(server.URL + path))

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

			fields := Fields{
				"FeedType": Equal(string(feedType)),
				"Title":    Equal("My Test Feed"),
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
			}

			switch feedType {
			case "rss":
				fields["Link"] = Equal("https://example.testing" + api.RSSPath)
			case "atom":
				fields["FeedLink"] = Equal("https://example.testing" + api.AtomPath)
			}

			Expect(feed).To(PointTo(MatchFields(IgnoreExtras, fields)))
		},

		Entry(
			"RSS feed",
			"rss",
			api.RSSPath,
			"application/rss+xml; charset=utf-8",
		),

		Entry("Atom feed",
			"atom",
			api.AtomPath,
			"application/atom+xml; charset=utf-8",
		),
	)
})

var _ = Describe("iCal feed output", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func(ctx SpecContext) {
		By("inserting events", func() {
			Expect(model.InsertEvent(ctx, db, event1)).To(Succeed())
			Expect(model.InsertEvent(ctx, db, event2)).To(Succeed())
			Expect(model.InsertEvent(ctx, db, event3)).To(Succeed())
		})

		e := echo.New()
		h := api.NewFeedHandler(db, api.Config{
			PageTitle: "My Test Feed",
			BaseURL:   Must(url.Parse("https://example.testing")),
		})
		h.Register(e)

		server = httptest.NewServer(e)
		DeferCleanup(server.Close)
	})

	Specify("iCal feed", func() {
		r := Must(server.Client().Get(server.URL + api.ICalPath))

		Expect(r.StatusCode).To(Equal(http.StatusOK))
		Expect(r.Header).To(SatisfyAll(
			HaveKeyWithValue(echo.HeaderContentType, HaveExactElements(
				Equal("text/calendar; charset=utf-8"),
			)),
			HaveKeyWithValue(echo.HeaderContentDisposition, HaveExactElements(
				Equal(`attachment; filename="calendar.ics"`),
			)),
		))

		cal := Must(ics.ParseCalendar(r.Body))

		Expect(cal.CalendarProperties).To(ContainElements(
			HaveField("BaseProperty", MatchFields(IgnoreExtras, Fields{
				"IANAToken": Equal("METHOD"),
				"Value":     Equal("PUBLISH"),
			})),
			HaveField("BaseProperty", MatchFields(IgnoreExtras, Fields{
				"IANAToken": Equal("DESCRIPTION"),
				"Value":     Equal("My Test Feed"),
			})),
			HaveField("BaseProperty", MatchFields(IgnoreExtras, Fields{
				"IANAToken": Equal("URL"),
				"Value":     Equal("https://example.testing" + api.ICalPath),
			})),
		))

		var matchers []any

		for _, target := range []*domain.Event{event3, event2, event1} {
			matchers = append(matchers, MakeMatcher(func(ev *ics.VEvent) (bool, error) {
				Expect(Must(ev.GetLastModifiedAt())).To(BeTemporally("~", time.Now(), time.Second))
				Expect(Must(ev.GetStartAt())).To(BeTemporally("~", target.StartAt.Time(), time.Second))

				if target.EndAt.Time().IsZero() {
					_, err := ev.GetEndAt()
					Expect(err).To(HaveOccurred())
				} else {
					Expect(Must(ev.GetEndAt())).To(BeTemporally("~", target.EndAt.Time(), time.Second))
				}

				summary := ev.GetProperty(ics.ComponentPropertySummary)
				Expect(summary).To(HaveField("Value", target.Title))

				url := ev.GetProperty(ics.ComponentPropertyUrl)
				Expect(url).To(HaveField("Value", target.URL))

				desc := ev.GetProperty(ics.ComponentPropertyDescription)
				Expect(desc).To(HaveField("Value", target.GetDescription()))

				return true, nil
			}))
		}

		Expect(cal.Events()).To(HaveExactElements(matchers...))
	})
})
