package api_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/api"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gcustom"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("iCal feed output", func() {
	var (
		rec *httptest.ResponseRecorder
		c   echo.Context
		h   *api.ICalHandler
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

		h = api.NewICalHandler(db, api.FeedConfig{
			Title: "My Test Feed",
			Link:  "https://example.testing",
		})
	})

	Specify("iCal feed", func() {
		Expect(h.Handle(c)).To(Succeed())

		r := rec.Result()
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
				"Value":     Equal("https://example.testing"),
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
