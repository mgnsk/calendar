package domain_test

import (
	"github.com/mgnsk/calendar/domain"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("creating a stop word list", func() {
	Specify("words are trimmed", func() {
		words := domain.NewStopWordList(" a ")

		Expect(words).To(HaveExactElements(
			"a",
		))
	})

	Specify("words are lowercased", func() {
		words := domain.NewStopWordList("A")

		Expect(words).To(HaveExactElements(
			"a",
		))
	})

	Specify("duplicate words are removed", func() {
		words := domain.NewStopWordList("A", "a")

		Expect(words).To(HaveExactElements(
			"a",
		))
	})
})
