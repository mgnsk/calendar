package model_test

import (
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("setting stopwords", func() {
	When("word does not exist", func() {
		It("is inserted", func(ctx SpecContext) {
			Expect(model.SetStopWords(ctx, db, domain.NewStopWordList("word1", "word2"))).To(Succeed())

			words := Must(model.ListStopWords(ctx, db))
			Expect(words).To(HaveExactElements(
				"word1",
				"word2",
			))
		})
	})

	When("word exists", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.SetStopWords(ctx, db, domain.NewStopWordList("word1", "word2"))).To(Succeed())
		})

		It("is ignored", func(ctx SpecContext) {
			Expect(model.SetStopWords(ctx, db, domain.NewStopWordList("word1", "word2", "word3"))).To(Succeed())

			words := Must(model.ListStopWords(ctx, db))

			Expect(words).To(HaveExactElements(
				"word1",
				"word2",
				"word3",
			))
		})
	})
})
