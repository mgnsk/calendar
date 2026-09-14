package model_test

import (
	"sync"
	"time"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("updating events", func() {
	var (
		ev *domain.Event
	)

	JustBeforeEach(func(ctx SpecContext) {
		ev = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
			Title:       "Old title",
			Description: "Old description",
			URL:         "https://old.testing",
			Location:    "old",
			OSMType:     "node",
			OSMID:       123,
			Latitude:    1,
			Longitude:   1,
			IsDraft:     false,
			UserID:      snowflake.Generate(),
		}

		Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())

		By("asserting tags are created", func() {
			tags := Must(model.ListTags(ctx, db, time.Time{}, time.Time{}, 0, 0))

			Expect(tags).To(HaveExactElements(
				HaveField("Name", "description"),
				HaveField("Name", "old"),
				HaveField("Name", "title"),
			))
		})
	})

	When("event is saved as a draft", func() {
		JustBeforeEach(func(ctx SpecContext) {
			ev.IsDraft = true
			Expect(model.UpdateEvent(ctx, db, ev)).To(Succeed())
		})

		Specify("tags are removed", func(ctx SpecContext) {
			tags := Must(model.ListTags(ctx, db, time.Time{}, time.Time{}, 0, 0))

			Expect(tags).To(BeEmpty())
		})
	})
})

var _ = Describe("listing events", func() {
	var (
		userID1 snowflake.ID
		userID2 snowflake.ID
	)

	JustBeforeEach(func(ctx SpecContext) {
		userID1 = snowflake.Generate()
		userID2 = snowflake.Generate()

		By("inserting events", func() {
			events := []*domain.Event{
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(3 * time.Hour),
					Title:       "Event 1",
					Description: "Desc 1",
					URL:         "",
					UserID:      userID1,
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(2 * time.Hour),
					Title:       "Event 2",
					Description: "Desc 2",
					URL:         "",
					UserID:      userID1,
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					Title:       "Event 3",
					Description: "Desc 3",
					URL:         "",
					UserID:      userID2,
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					Title:       "Event 4",
					Description: "Desc 4",
					URL:         "",
					IsDraft:     true,
					UserID:      userID2,
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	Specify("draft event tags are not inserted", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db, time.Time{}, time.Time{}, 0, 0))

		Expect(tags).To(HaveExactElements(
			PointTo(MatchAllFields(Fields{
				"Name":       Equal("desc"),
				"EventCount": Equal(uint64(3)),
			})),
			PointTo(MatchAllFields(Fields{
				"Name":       Equal("event"),
				"EventCount": Equal(uint64(3)),
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
					Title:       "Event 1",
					Description: "Desc 1",
					URL:         "",
					UserID:      snowflake.Generate(),
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     startTime,
					Title:       "Event ÕÄÖÜ 😀😀😀",
					Description: "Desc 2 some@email.testing, https://outlink.testing",
					URL:         "",
					UserID:      snowflake.Generate(),
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					Title:       "Event 3",
					Description: "Desc 3",
					URL:         "",
					UserID:      snowflake.Generate(),
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	DescribeTable("incorrect queries",
		func(ctx SpecContext, query string) {
			result, err := model.NewEventsQuery().
				WithStartAtFrom(time.Now().Add(1*time.Hour).Add(30*time.Minute)).
				WithStartAtUntil(time.Now().Add(2*time.Hour).Add(30*time.Minute)).
				WithOrder(0, model.OrderStartAtAsc).
				WithSearchText(query).
				List(ctx, db)

			Expect(result).To(BeEmpty())
			Expect(err).To(MatchError(calendar.InvalidValue))
		},
		Entry("only AND operator", "AND"),
		Entry("multiple operators prefix", "AND AND Desc"),
		Entry("multiple operators suffix", "Desc AND AND"),
		Entry("unused AND operator", "AND something"),
		Entry("syntax error", `a"Desc"a`),
	)

	DescribeTable("valid queries",
		func(ctx SpecContext, query string) {
			result := Must(
				model.NewEventsQuery().
					WithStartAtFrom(startTime).
					WithStartAtUntil(endTime).
					WithOrder(0, model.OrderStartAtAsc).
					WithSearchText(query).
					List(ctx, db),
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
		Entry("AND operator", `"Desc 2" AND ÕÄÖÜ`),
		Entry("NOT operator", `"Desc 2" NOT "Desc 3"`),
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
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "",
				UserID:      snowflake.Generate(),
			}

			wg.Go(func() {
				defer GinkgoRecover()

				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			})
		}

		wg.Wait()

		events := Must(model.NewEventsQuery().WithOrder(0, model.OrderStartAtAsc).List(ctx, db))
		Expect(events).To(HaveLen(100))
	})
})
