package model_test

import (
	"sync"
	"time"

	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	"github.com/mgnsk/calendar/pkg/wreck"
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
			ev = &domain.Event{
				ID:          snowflake.Generate(),
				StartAt:     time.Now().Add(2 * time.Hour),
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "https://calendar.testing",
				Location:    "hash",
				IsDraft:     false,
				UserID:      snowflake.Generate(),
			}

			Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
		})

		Specify("event is persisted", func(ctx SpecContext) {
			By("asserting event can be retrieved", func() {
				event := Must(model.GetEvent(ctx, db, ev.ID))

				Expect(event).To(SatisfyAll(
					HaveField("GetCreatedAt()", BeTemporally("~", time.Now(), time.Second)),
					PointTo(MatchAllFields(Fields{
						"ID":          Equal(ev.ID),
						"StartAt":     BeTemporally("~", ev.StartAt, time.Second),
						"Title":       Equal(ev.Title),
						"Description": Equal(ev.Description),
						"URL":         Equal(ev.URL),
						"Location":    Equal("hash"),
						"IsDraft":     BeFalse(),
						"UserID":      Equal(ev.UserID),
					})),
				))
			})

			By("asserting event can be listed", func() {
				result := Must(model.NewEventsQuery().WithOrder(0, model.OrderStartAtAsc).List(ctx, db, false, ""))

				Expect(result).To(HaveExactElements(
					SatisfyAll(
						HaveField("GetCreatedAt()", BeTemporally("~", time.Now(), time.Second)),
						PointTo(MatchAllFields(Fields{
							"ID":          Equal(ev.ID),
							"StartAt":     BeTemporally("~", ev.StartAt, time.Second),
							"Title":       Equal(ev.Title),
							"Description": Equal(ev.Description),
							"URL":         Equal(ev.URL),
							"Location":    Equal("hash"),
							"IsDraft":     BeFalse(),
							"UserID":      Equal(ev.UserID),
						})),
					),
				))
			})
		})
	})
})

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
			IsDraft:     false,
			UserID:      snowflake.Generate(),
		}

		Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())

		By("asserting tags are created", func() {
			tags := Must(model.ListTags(ctx, db, 0))

			Expect(tags).To(HaveExactElements(
				HaveField("Name", "description"),
				HaveField("Name", "old"),
				HaveField("Name", "title"),
			))
		})
	})

	Specify("event can be updated", func(ctx SpecContext) {
		ev.Title = "New title"
		ev.Description = "New description"
		ev.URL = "https://new.testing"
		ev.StartAt = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		ev.Location = "new"
		ev.IsDraft = true

		Expect(model.UpdateEvent(ctx, db, ev)).To(Succeed())

		By("asserting updated event was persisted", func() {
			event := Must(model.GetEvent(ctx, db, ev.ID))

			Expect(event).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"Title":       Equal("New title"),
				"Description": Equal("New description"),
				"URL":         Equal("https://new.testing"),
				"StartAt":     BeTemporally("~", ev.StartAt),
				"Location":    Equal("new"),
				"IsDraft":     BeTrue(),
			})))
		})

		By("asserting tags are updated", func() {
			tags := Must(model.ListTags(ctx, db, 0))

			Expect(tags).To(HaveExactElements(
				HaveField("Name", "description"),
				HaveField("Name", "new"),
				HaveField("Name", "title"),
			))
		})
	})
})

var _ = Describe("deleting events", func() {
	var (
		ev *domain.Event
	)

	JustBeforeEach(func(ctx SpecContext) {
		ev = &domain.Event{
			ID:          snowflake.Generate(),
			StartAt:     time.Now().Add(2 * time.Hour),
			Title:       "Old title",
			Description: "Old description",
			URL:         "",
			UserID:      snowflake.Generate(),
		}

		Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())

		By("asserting tags are created", func() {
			tags := Must(model.ListTags(ctx, db, 0))

			Expect(tags).To(HaveExactElements(
				HaveField("Name", "description"),
				HaveField("Name", "old"),
				HaveField("Name", "title"),
			))
		})
	})

	Specify("event can be deleted", func(ctx SpecContext) {
		Expect(model.DeleteEvent(ctx, db, ev)).To(Succeed())

		By("asserting event was deleted", func() {
			Expect(model.GetEvent(ctx, db, ev.ID)).Error().To(MatchError(wreck.NotFound))
		})

		By("asserting tags are updated", func() {
			tags := Must(model.ListTags(ctx, db, 0))
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

	Specify("events can be filtered by user", func(ctx SpecContext) {
		result := Must(
			model.NewEventsQuery().
				WithOrder(0, model.OrderStartAtAsc).
				WithUserID(userID1).
				List(ctx, db, false, ""),
		)

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
			})),
		))
	})

	Specify("draft events can be included", func(ctx SpecContext) {
		result := Must(
			model.NewEventsQuery().
				WithOrder(0, model.OrderStartAtAsc).
				List(ctx, db, true, ""),
		)

		Expect(result).To(HaveLen(4))
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
		Entry("multiple operators prefix", "AND AND Desc"),
		Entry("multiple operators suffix", "Desc AND AND"),
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
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "",
				UserID:      snowflake.Generate(),
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
