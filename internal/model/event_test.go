package model_test

import (
	"sync"
	"time"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
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
			result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderStartAtAsc, 0, 0))

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
						"TagRelations": HaveExactElements(
							HaveField("EventCount", uint64(1)),
							HaveField("EventCount", uint64(1)),
						),
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
					Tags:        []string{"tag2", "tag3"},
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	Specify("events can be listed in start time order ascending", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderStartAtAsc, 0, 0))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag3"),
						"EventCount": Equal(uint64(1)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
		))
	})

	Specify("events can be listed in start time order descending", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderStartAtDesc, 0, 0))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag3"),
						"EventCount": Equal(uint64(1)),
					})),
				),
			})),
		))
	})

	Specify("events can be listed in created at time order descending", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderCreatedAtDesc, 0, 0))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag3"),
						"EventCount": Equal(uint64(1)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
		))
	})

	Specify("events can be filtered by time", func(ctx SpecContext) {
		result := Must(model.ListEvents(
			ctx,
			db,
			time.Now().Add(1*time.Hour).Add(30*time.Minute),
			time.Now().Add(2*time.Hour).Add(30*time.Minute),
			"",
			model.OrderStartAtAsc,
			0,
			0,
		))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
		))
	})

	Specify("events can be filtered by tags", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderStartAtAsc, 0, 0, "tag1"))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
		))
	})

	Specify("empty tag filter is skipped", func(ctx SpecContext) {
		result := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderStartAtAsc, 0, 0, ""))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 3"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag3"),
						"EventCount": Equal(uint64(1)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 1"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
				),
			})),
		))
	})

	Specify("events can be filtered by time and tags", func(ctx SpecContext) {
		result := Must(model.ListEvents(
			ctx,
			db,
			time.Now().Add(1*time.Hour).Add(30*time.Minute),
			time.Now().Add(2*time.Hour).Add(30*time.Minute),
			"",
			model.OrderStartAtAsc,
			0,
			0,
			"tag1",
		))

		Expect(result).To(HaveExactElements(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"Title": Equal("Event 2"),
				"TagRelations": HaveExactElements(
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag1"),
						"EventCount": Equal(uint64(2)),
					})),
					PointTo(MatchFields(IgnoreExtras, Fields{
						"Name":       Equal("tag2"),
						"EventCount": Equal(uint64(2)),
					})),
				),
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
					StartAt:     timestamp.New(time.Now().Add(3 * time.Hour)),
					EndAt:       timestamp.Timestamp{},
					Title:       "Event 1",
					Description: "Desc 1",
					URL:         "",
					Tags:        []string{"tag1"},
				},
				{
					ID:          snowflake.Generate(),
					StartAt:     timestamp.New(startTime),
					EndAt:       timestamp.Timestamp{},
					Title:       "Event ÕÄÖÜ 😀",
					Description: "Desc 2 some@email.testing, https://outlink.testing",
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
					Tags:        []string{"tag2", "tag3"},
				},
			}

			for _, ev := range events {
				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}
		})
	})

	DescribeTable("incorrect queries",
		func(ctx SpecContext, query string) {
			_, err := model.ListEvents(
				ctx,
				db,
				time.Now().Add(1*time.Hour).Add(30*time.Minute),
				time.Now().Add(2*time.Hour).Add(30*time.Minute),
				query,
				model.OrderStartAtAsc,
				0,
				0,
				"tag1",
			)

			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(wreck.NotFound))
		},
		Entry("emoji", `😀`),
		Entry("invalid utf8", `😀`[:len(`😀`)-1]),
		Entry("backslash", `aou\`),
		Entry("wildcard in beginning", `*aou`),
		Entry("wildcard in the end", `aou*`), // Note: we quote searches, making this otherwise valid query invalid.
		Entry("no multiple exact match", `"Desc 2" "unknown@email.testing"`),
	)

	DescribeTable("valid queries",
		func(ctx SpecContext, query string) {
			result := Must(model.ListEvents(
				ctx,
				db,
				startTime,
				endTime,
				query,
				model.OrderStartAtAsc,
				0,
				0,
				"tag1",
			))

			Expect(result).To(HaveExactElements(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Title":       Equal("Event ÕÄÖÜ 😀"),
					"Description": HavePrefix("Desc 2"),
					"TagRelations": HaveExactElements(
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":       Equal("tag1"),
							"EventCount": Equal(uint64(2)),
						})),
						PointTo(MatchFields(IgnoreExtras, Fields{
							"Name":       Equal("tag2"),
							"EventCount": Equal(uint64(2)),
						})),
					),
				})),
			))
		},
		Entry("letters", `aou`),
		Entry("multi word exact match", `Desc 2`),
		Entry("exact match", `"Desc 2"`),
		Entry("spaces", "Desc \t \u00a0  3"),
		Entry("invalid", "Desc \t \xa0  3"),
		Entry("special characters", `äöü`),
		Entry("partial word", `des`),
		Entry("partial word no prefix", `esc`),
		Entry("partial words", `des even`),
		Entry("partial word", `even`),
		Entry("multiple exact match", `"Desc 2" "some@email.testing"`),
		Entry("email", `some@email.testing`),
		Entry("day", "3rd"),
		Entry("day and month", "3 jan"),
		Entry("month and day", "jan 3"),
		Entry("day and partial month", "3 ja"),
	)
})

var _ = Describe("concurrent insert", func() {
	Specify("test", func(ctx SpecContext) {
		concurrency := 100
		wg := sync.WaitGroup{}

		for range concurrency {
			ev := &domain.Event{
				ID:          snowflake.Generate(),
				StartAt:     timestamp.New(time.Now().Add(2 * time.Hour)),
				EndAt:       timestamp.Timestamp{},
				Title:       "Event Title ÕÄÖÜ 1",
				Description: "Desc 1",
				URL:         "",
				Tags:        []string{"tag1", "tag2"},
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()

				Expect(model.InsertEvent(ctx, db, ev)).To(Succeed())
			}()
		}

		wg.Wait()

		events := Must(model.ListEvents(ctx, db, time.Time{}, time.Time{}, "", model.OrderStartAtAsc, 0, 0))
		Expect(events).To(HaveLen(100))
	})
})
