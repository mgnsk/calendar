package model_test

import (
	"sync"
	"time"

	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting events", func() {
	When("event is inserted", func() {
		var (
			ev *domain.Event
		)

		JustBeforeEach(func(ctx SpecContext) {
			By("inserting one existing tag", func() {
				Expect(model.InsertTags(ctx, db, "tag1")).To(Succeed())
			})

			ev = &domain.Event{
				ID:          snowflake.Generate(),
				StartAt:     time.Now().Add(2 * time.Hour),
				EndAt:       time.Time{},
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "",
			}

			Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
		})

		Specify("event is persisted", func(ctx SpecContext) {
			result := Must(model.NewEventsQuery().WithOrder(0, model.OrderStartAtAsc).List(ctx, db, false, ""))

			Expect(result).To(HaveExactElements(
				SatisfyAll(
					HaveField("GetCreatedAt()", BeTemporally("~", time.Now(), time.Second)),
					PointTo(MatchAllFields(Fields{
						"ID":          Equal(ev.ID),
						"StartAt":     BeTemporally("~", ev.StartAt, time.Second),
						"EndAt":       BeZero(),
						"Title":       Equal(ev.Title),
						"Description": Equal(ev.Description),
						"URL":         Equal(ev.URL),
						"IsDraft":     BeFalse(),
					})),
				),
			))
		})
	})
})

var _ = Describe("listing events", func() {
	JustBeforeEach(func(ctx SpecContext) {
		By("inserting events", func() {
			events := []*domain.Event{
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(3 * time.Hour),
					EndAt:       time.Time{},
					Title:       "Event 1",
					Description: "Desc 1",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(2 * time.Hour),
					EndAt:       time.Time{},
					Title:       "Event 2",
					Description: "Desc 2",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					EndAt:       time.Now().Add(2 * time.Hour),
					Title:       "Event 3",
					Description: "Desc 3",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					EndAt:       time.Now().Add(2 * time.Hour),
					Title:       "Event 4",
					Description: "Desc 4",
					URL:         "",
					IsDraft:     true,
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	Specify("events can be listed in start time order ascending", func(ctx SpecContext) {
		result := Must(model.NewEventsQuery().WithOrder(0, model.OrderStartAtAsc).List(ctx, db, false, ""))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
			})),
		))
	})

	Specify("events can be listed in start time order descending", func(ctx SpecContext) {
		result := Must(model.NewEventsQuery().WithOrder(0, model.OrderStartAtDesc).List(ctx, db, false, ""))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
			})),
		))
	})

	Specify("events can be listed in created at time order descending", func(ctx SpecContext) {
		result := Must(model.NewEventsQuery().WithOrder(0, model.OrderCreatedAtDesc).List(ctx, db, false, ""))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
			})),
		))
	})

	Specify("events can be filtered by time", func(ctx SpecContext) {
		result := Must(
			model.NewEventsQuery().
				WithStartAtFrom(time.Now().Add(1*time.Hour).Add(30*time.Minute)).
				WithStartAtUntil(time.Now().Add(2*time.Hour).Add(30*time.Minute)).
				WithOrder(0, model.OrderStartAtAsc).
				List(ctx, db, false, ""),
		)

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
			})),
		))
	})
})

var _ = Describe("full text search", func() {
	var (
		startTime, endTime time.Time
	)

	JustBeforeEach(func(ctx SpecContext) {
		startTime = Must(time.Parse(time.RFC3339, "2025-01-03T18:00:00+02:00"))
		endTime = startTime.Add(time.Hour)

		By("inserting events", func() {
			events := []*domain.Event{
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(3 * time.Hour),
					EndAt:       time.Time{},
					Title:       "Event 1",
					Description: "Desc 1",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     startTime,
					EndAt:       time.Time{},
					Title:       "Event ÕÄÖÜ 😀😀😀",
					Description: "Desc 2 some@email.testing, https://outlink.testing",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					EndAt:       time.Now().Add(2 * time.Hour),
					Title:       "Event 3",
					Description: "Desc 3",
					URL:         "",
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	DescribeTable("incorrect queries",
		func(ctx SpecContext, query string) {
			result := Must(
				model.NewEventsQuery().
					WithStartAtFrom(time.Now().Add(1*time.Hour).Add(30*time.Minute)).
					WithStartAtUntil(time.Now().Add(2*time.Hour).Add(30*time.Minute)).
					WithOrder(0, model.OrderStartAtAsc).
					List(ctx, db, false, query),
			)

			Expect(result).To(BeEmpty())
		},
		Entry("multiple exact match at least one", `"Desc 2" "unknown@email.testing"`), // Defaults to AND operator.
		Entry("only AND operator", "AND"),
		Entry("unused AND operator", "AND something"),
		Entry("backslash", `aou\`),
		Entry("spaces", "Desc \t \u00a0  3"),
		Entry("wildcard in beginning", `*aou`),
		Entry("wildcard in the end", `Des*`), // Note: we quote searches, making this otherwise valid query invalid.
	)

	DescribeTable("valid queries",
		func(ctx SpecContext, query string) {
			result := Must(
				model.NewEventsQuery().
					WithStartAtFrom(startTime).
					WithStartAtUntil(endTime).
					WithOrder(0, model.OrderStartAtAsc).
					List(ctx, db, false, query),
			)

			Expect(result).To(HaveExactElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Title":       Equal("Event ÕÄÖÜ 😀😀😀"),
					"Description": HavePrefix("Desc 2"),
				})),
			))
		},
		Entry("letters", `aou`),
		Entry("emoji", `😀😀😀`),
		Entry("multi word exact match", `Desc 2`),
		Entry("quoted exact match", `"Desc 2"`),
		Entry("exact match", `"Desc 2"`),
		Entry("special characters", `äöü`),
		Entry("partial word", `des`),
		Entry("partial word no prefix", `esc`),
		Entry("partial words", `des even`),
		Entry("partial word", `even`),
		Entry("multiple exact match", `"Desc 2" "some@email.testing"`),
		Entry("OR operator", `"Desc 2" OR "some@email.testing"`),
		Entry("email", `some@email.testing`),
	)
})

var _ = Describe("concurrent insert", func() {
	Specify("test", func(ctx SpecContext) {
		concurrency := 100
		wg := sync.WaitGroup{}

		for range concurrency {
			ev := &domain.Event{
				ID:          snowflake.Generate(),
				StartAt:     time.Now().Add(2 * time.Hour),
				EndAt:       time.Time{},
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "",
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}()
		}

		wg.Wait()

		events := Must(model.NewEventsQuery().WithOrder(0, model.OrderStartAtAsc).List(ctx, db, false, ""))
		Expect(events).To(HaveLen(100))
	})
})
