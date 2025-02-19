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

var _ = Describe("inserting tags", func() {
	When("tag does not exist", func() {
		It("is inserted", func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2")).To(Succeed())

			tags := Must(model.ListTags(ctx, db, 0))
			Expect(tags).To(HaveExactElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("tag1"),
				})),

				PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("tag2"),
				})),
			))
		})
	})

	When("tag exists", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2")).To(Succeed())
		})

		It("is ignored", func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2", "tag3")).To(Succeed())

			tags := Must(model.ListTags(ctx, db, 0))

			Expect(tags).To(HaveExactElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("tag1"),
				})),

				PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("tag2"),
				})),

				PointTo(MatchFields(IgnoreExtras, Fields{
					"Name": Equal("tag3"),
				})),
			))
		})
	})
})

var _ = Describe("listing tags", func() {
	JustBeforeEach(func(ctx SpecContext) {
		By("inserting events", func() {
			events := []*domain.Event{
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(3 * time.Hour),
					EndAt:       time.Time{},
					Title:       "Event 1",
					Description: "Desc 1 tag1",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(2 * time.Hour),
					EndAt:       time.Time{},
					Title:       "Event 2",
					Description: "Desc 2 tag1 tag2",
					URL:         "",
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     time.Now().Add(1 * time.Hour),
					EndAt:       time.Now().Add(2 * time.Hour),
					Title:       "Event 3",
					Description: "Desc 3 tag3",
					URL:         "",
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	Specify("tags contain the number of related events", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db, 0))

		Expect(tags).To(ContainElements(
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
})
