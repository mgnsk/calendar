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
				StartAt:     timestamp.New(time.Now().Add(2 * time.Hour)),
				EndAt:       timestamp.Timestamp{},
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "",
				Tags:        []string{"tag1", "tag2"},
			}

			Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
		})

		Specify("event is persisted", func(ctx SpecContext) {
			result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "asc"))

			Expect(result).To(HaveExactElements(
				SatisfyAll(
					HaveField("GetCreatedAt()", BeTemporally("~", time.Now(), time.Second)),
					PointTo(MatchAllFields(Fields{
						"ID":          Equal(ev.ID),
						"StartAt":     HaveField("Time()", BeTemporally("~", ev.StartAt.Time(), time.Second)),
						"EndAt":       HaveField("Time()", BeZero()),
						"Title":       Equal(ev.Title),
						"Description": Equal(ev.Description),
						"URL":         Equal(ev.URL),
						"Tags":        HaveExactElements("tag1", "tag2"),
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

	Specify("events can be listed in start time order ascending", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "asc"))

		Expect(result).To(HaveExactElements(
			HaveField("Title", "Event 3"),
			HaveField("Title", "Event 2"),
			HaveField("Title", "Event 1"),
		))
	})

	Specify("events can be listed in start time order descending", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "desc"))

		Expect(result).To(HaveExactElements(
			HaveField("Title", "Event 1"),
			HaveField("Title", "Event 2"),
			HaveField("Title", "Event 3"),
		))
	})

	Specify("events can be filtered by time", func(ctx SpecContext) {
		result := Must(model.ListEvents(
			ctx,
			db,
			time.Now().Add(1*time.Hour).Add(30*time.Minute),
			time.Now().Add(2*time.Hour).Add(30*time.Minute),
			"asc",
		))

		Expect(result).To(HaveExactElements(
			HaveField("Title", "Event 2"),
		))
	})

	Specify("events can be filtered by tags", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "asc", "tag1"))

		Expect(result).To(HaveExactElements(
			HaveField("Title", "Event 2"),
			HaveField("Title", "Event 1"),
		))
	})

	Specify("events can be filtered by time and tags", func(ctx SpecContext) {
		result := Must(model.ListEvents(
			ctx,
			db,
			time.Now().Add(1*time.Hour).Add(30*time.Minute),
			time.Now().Add(2*time.Hour).Add(30*time.Minute),
			"asc",
			"tag1",
		))

		Expect(result).To(HaveExactElements(
			HaveField("Title", "Event 2"),
		))
	})
})
