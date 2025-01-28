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

			tags, err := model.ListTags(ctx, db)
			Expect(err).NotTo(HaveOccurred())

			Expect(tags).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
					"Name": Equal("tag1"),
				})),
			))
		})
	})

	When("tag exists", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())
		})

		It("is ignored", func(ctx SpecContext) {
			Expect(model.InsertTag(ctx, db, "tag1")).To(Succeed())

			tags, err := model.ListTags(ctx, db)
			Expect(err).NotTo(HaveOccurred())

			Expect(tags).To(ConsistOf(
				PointTo(MatchFields(IgnoreExtras, Fields{
					"ID":   Not(BeZero()),
					"Name": Equal("tag1"),
				})),
			))
		})
	})
})
