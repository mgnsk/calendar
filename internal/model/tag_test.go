package model_test

import (
	"time"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting tags", func() {
	When("tag does not exist", func() {
		It("is inserted", func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2")).To(Succeed())

			tags := Must(model.ListTags(ctx, db))
			Expect(tags).To(HaveExactElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
					"Name": Equal("tag1"),
				})),

				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
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

			tags := Must(model.ListTags(ctx, db))

			Expect(tags).To(HaveExactElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
					"Name": Equal("tag1"),
				})),

				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
					"Name": Equal("tag2"),
				})),

				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
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
					StartAt:     timestamp.New(time.Now().Add(3 * time.Hour)),
					EndAt:       timestamp.Timestamp{},
					Title:       "Event 1",
					Description: "Desc 1",
					URL:         "",
					Tags:        []string{"tag1"},
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     timestamp.New(time.Now().Add(2 * time.Hour)),
					EndAt:       timestamp.Timestamp{},
					Title:       "Event 2",
					Description: "Desc 2",
					URL:         "",
					Tags:        []string{"tag1", "tag2"},
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     timestamp.New(time.Now().Add(1 * time.Hour)),
					EndAt:       timestamp.New(time.Now().Add(2 * time.Hour)),
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
	})

	Specify("tags contain the number of related events", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db))

		Expect(tags).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":         Not(BeZero()),
				"Name":       Equal("tag1"),
				"EventCount": Equal(uint64(2)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":         Not(BeZero()),
				"Name":       Equal("tag2"),
				"EventCount": Equal(uint64(1)),
			})),

			PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":         Not(BeZero()),
				"Name":       Equal("tag3"),
				"EventCount": Equal(uint64(1)),
			})),
		))
	})
})
