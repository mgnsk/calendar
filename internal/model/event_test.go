package model_test

import (
	"time"

	"github.com/mgnsk/calendar/internal/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting events", func() {
	Context("inserting events without tags", func() {
		Specify("event is persisted", func(ctx SpecContext) {
			ts := time.Now()

			By("inserting events", func() {
				Expect(model.InsertEvent(ctx, db, ts.Add(2*time.Hour), "Event Title ÕÄÖÜ 1", "Content 1")).To(Succeed())
			})

			events, err := model.ListEvents(ctx, db, "", time.Time{}, time.Time{}, 0)
			Expect(err).NotTo(HaveOccurred())

			Expect(events).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Tags": BeNil(),
				})),
			))
		})
	})

	Context("inserting event with tags", func() {
		var tags []*model.Tag

		JustBeforeEach(func(ctx SpecContext) {
			By("creating tags", func() {
				Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())
				Expect(model.InsertTag(ctx, db, "tag2")).To(Succeed())

				m, err := model.ListTags(ctx, db)
				Expect(err).NotTo(HaveOccurred())
				Expect(m).To(HaveLen(2))

				tags = m
			})
		})

		Specify("event is persisted", func(ctx SpecContext) {
			ts := time.Now()

			By("inserting events", func() {
				Expect(model.InsertEvent(ctx, db, ts.Add(2*time.Hour), "Event Title ÕÄÖÜ 1", "Content 1", tags...)).To(Succeed())
			})

			events, err := model.ListEvents(ctx, db, "", time.Time{}, time.Time{}, 0)
			Expect(err).NotTo(HaveOccurred())

			Expect(events).To(ConsistOf(
				HaveField("Tags", ConsistOf(
					HaveField("Name", "tag1"),
					HaveField("Name", "tag2"),
				)),
			))
		})
	})
})

var _ = Describe("listing events", func() {
	Specify("events are ordered by future events first", func(ctx SpecContext) {
		ts := time.Now()

		By("inserting events", func() {
			Expect(model.InsertEvent(ctx, db, ts.Add(time.Hour), "Event Title ÕÄÖÜ 2", "Content 2")).To(Succeed())
			Expect(model.InsertEvent(ctx, db, ts.Add(2*time.Hour), "Event Title ÕÄÖÜ 1", "Content 1")).To(Succeed())
		})

		events, err := model.ListEvents(ctx, db, "", time.Time{}, time.Time{}, 0)
		Expect(err).NotTo(HaveOccurred())

		Expect(events).To(ConsistOf(
			PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":               Not(BeZero()),
				"UnixTimestamp":    Equal(ts.Add(2 * time.Hour).Unix()),
				"RFC3339Timestamp": Equal(ts.Add(2 * time.Hour).Format(time.RFC3339)),
				"Title":            Equal("Event Title ÕÄÖÜ 1"),
				"Content":          Equal("Content 1"),
				"Tags":             BeNil(),
			})),
			PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":               Not(BeZero()),
				"UnixTimestamp":    Equal(ts.Add(time.Hour).Unix()),
				"RFC3339Timestamp": Equal(ts.Add(time.Hour).Format(time.RFC3339)),
				"Title":            Equal("Event Title ÕÄÖÜ 2"),
				"Content":          Equal("Content 2"),
				"Tags":             BeNil(),
			})),
		))
	})

	When("events are filtered by tag", func() {
		var tags []*model.Tag

		JustBeforeEach(func(ctx SpecContext) {
			By("creating tags", func() {
				Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())
				Expect(model.InsertTag(ctx, db, "tag2")).To(Succeed())

				m, err := model.ListTags(ctx, db)
				Expect(err).NotTo(HaveOccurred())
				Expect(m).To(HaveLen(2))

				tags = m
			})

			By("creating multiple events", func() {
				Expect(model.InsertEvent(ctx, db, time.Now(), "Event1", "", tags[0]))
				Expect(model.InsertEvent(ctx, db, time.Now().Add(time.Hour), "Event2", "", tags[1]))
				Expect(model.InsertEvent(ctx, db, time.Now().Add(2*time.Hour), "Event3", "", tags...))
			})
		})

		Specify("matching events are returned", func(ctx SpecContext) {
			events, err := model.ListEvents(ctx, db, "tag1", time.Time{}, time.Time{}, 0)
			Expect(err).NotTo(HaveOccurred())

			Expect(events).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Title": Equal("Event3"),
					"Tags": ConsistOf(
						HaveField("Name", "tag1"),
						HaveField("Name", "tag2"),
					),
				})),
				PointTo(MatchFields(IgnoreExtras, Fields{
					"Title": Equal("Event1"),
					"Tags": ConsistOf(
						HaveField("Name", "tag1"),
					),
				})),
			))
		})
	})
})
