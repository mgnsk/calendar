package model_test

import (
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Note: the happy-path insert flow is covered by the settings HTTP
// integration test (handler/settings_test.go). This case remains because it
// is an internal idempotency guarantee of SetStopWords not distinguishable
// from a single insert at the HTTP layer.
var _ = Describe("setting stopwords", func() {
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
