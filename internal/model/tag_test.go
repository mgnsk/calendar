package model_test

import (
	"github.com/mgnsk/calendar/internal/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting tags", func() {
	When("tag does not exist", func() {
		It("is inserted", func(ctx SpecContext) {
			Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())

			tag, err := model.GetTag(ctx, db, "tag1")
			Expect(err).NotTo(HaveOccurred())

			Expect(tag).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag1"),
			})))
		})
	})

	When("tag exists", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())
		})

		It("is ignored", func(ctx SpecContext) {
			Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())

			tag, err := model.GetTag(ctx, db, "tag1")
			Expect(err).NotTo(HaveOccurred())

			Expect(tag).To(PointTo(MatchFields(IgnoreExtras, Fields{
				"ID":   Not(BeZero()),
				"Name": Equal("tag1"),
			})))

			By("asserting that a single tag exists", func() {
				tags, err := model.ListTags(ctx, db, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(tags).To(HaveLen(1))
			})
		})
	})
})

var _ = Describe("listing tags", func() {
	JustBeforeEach(func(ctx SpecContext) {
		for _, tag := range []string{"tag1", "tag2", "other"} {
			Expect(model.InsertTag(ctx, db, tag)).To(Succeed())
		}
	})

	Specify("tags can be filtered", func(ctx SpecContext) {
		tags, err := model.ListTags(ctx, db, "tag")
		Expect(err).NotTo(HaveOccurred())

		Expect(tags).To(HaveExactElements(
			HaveField("Name", "tag1"),
			HaveField("Name", "tag2"),
		))
	})
})
