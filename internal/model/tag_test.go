package model_test

import (
	"github.com/mgnsk/calendar/internal/model"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting tags", func() {
	When("tag does not exist", func() {
		It("is inserted", func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2")).To(Succeed())

			tag1 := Must(model.GetTag(ctx, db, "tag1"))
			tag2 := Must(model.GetTag(ctx, db, "tag2"))

			Expect(tag1).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag1"),
			})))

			Expect(tag2).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag2"),
			})))
		})
	})

	When("tag exists", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2")).To(Succeed())
		})

		It("is ignored", func(ctx SpecContext) {
			Expect(model.InsertTags(ctx, db, "tag1", "tag2", "tag3")).To(Succeed())

			tag1 := Must(model.GetTag(ctx, db, "tag1"))
			tag2 := Must(model.GetTag(ctx, db, "tag2"))
			tag3 := Must(model.GetTag(ctx, db, "tag3"))

			Expect(tag1).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag1"),
			})))

			Expect(tag2).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag2"),
			})))

			Expect(tag3).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag3"),
			})))

			By("asserting that correct number of tags exists", func() {
				tags := Must(model.ListTags(ctx, db, ""))
				Expect(tags).To(HaveLen(3))
			})
		})
	})
})

var _ = Describe("listing tags", func() {
	JustBeforeEach(func(ctx SpecContext) {
		Expect(model.InsertTags(ctx, db, "tag1", "tag2", "other")).To(Succeed())
	})

	Specify("tags can be filtered", func(ctx SpecContext) {
		tags := Must(model.ListTags(ctx, db, "tag"))

		Expect(tags).To(HaveExactElements(
			HaveField("Name", "tag1"),
			HaveField("Name", "tag2"),
		))
	})
})
