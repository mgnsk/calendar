package model_test

import (
	"time"

	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("listing tags", func() {
	JustBeforeEach(func(ctx SpecContext) {
		By("inserting events", func() {
			events := []*domain.Event{
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(3 * time.Hour),
					Title:       "Event 1",
					Description: "Desc 1 tag1",
					URL:         "",
					UserID:      1,
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(2 * time.Hour),
					Title:       "Event 2",
					Description: "Desc 2 tag1 tag2",
					URL:         "",
					UserID:      1,
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					Title:       "Event 3",
					Description: "Desc 3 tag3",
					URL:         "",
					UserID:      2,
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(-24 * time.Hour),
					Title:       "Event 4",
					Description: "Desc 4 tag4",
					URL:         "",
					UserID:      2,
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	Specify("tags contain the number of related future events", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db, time.Now(), time.Time{}, 0, 0))

		Expect(tags).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("desc"),
				"EventCount": Equal(uint64(3)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("event"),
				"EventCount": Equal(uint64(3)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag1"),
				"EventCount": Equal(uint64(2)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag2"),
				"EventCount": Equal(uint64(1)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag3"),
				"EventCount": Equal(uint64(1)),
			})),
		))
	})

	Specify("tags contain the number of related past events", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db, time.Time{}, time.Now(), 0, 0))

		Expect(tags).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("desc"),
				"EventCount": Equal(uint64(1)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("event"),
				"EventCount": Equal(uint64(1)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag4"),
				"EventCount": Equal(uint64(1)),
			})),
		))
	})

	Specify("tags contain the number of related user events", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db, time.Time{}, time.Time{}, 1, 0))

		Expect(tags).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("desc"),
				"EventCount": Equal(uint64(2)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("event"),
				"EventCount": Equal(uint64(2)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag1"),
				"EventCount": Equal(uint64(2)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag2"),
				"EventCount": Equal(uint64(1)),
			})),
		))
	})

	Specify("tags exclude stopwords case-insensitive", func(ctx SpecContext) {
		Expect(model.SetStopWords(ctx, db, domain.NewStopWordList("desc", "TAG2"))).To(Succeed())

		tags := Must(model.ListTags(ctx, db, time.Now(), time.Time{}, 0, 0))

		Expect(tags).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("event"),
				"EventCount": Equal(uint64(3)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag1"),
				"EventCount": Equal(uint64(2)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"Name":       Equal("tag3"),
				"EventCount": Equal(uint64(1)),
			})),
		))
	})
})
