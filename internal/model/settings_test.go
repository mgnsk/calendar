package model_test

import (
	"net/url"

	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	. "github.com/mgnsk/calendar/internal/pkg/testing"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting settings", func() {
	When("settings don't exist", func() {
		Specify("settings are inserted", func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: false,
				Title:         "Page Title",
				Description:   "Description",
				BaseURL:       Must(url.Parse("https://events.testing")),
				SessionSecret: []byte("sess_secret"),
			})).To(Succeed())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings).To(PointTo(MatchAllFields(Fields{
				"IsInitialized": BeFalse(),
				"Title":         Equal("Page Title"),
				"Description":   Equal("Description"),
				"BaseURL":       HaveField("String()", "https://events.testing"),
				"SessionSecret": Equal([]byte("sess_secret")),
			})))
		})
	})

	When("settings exist", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: true,
				Title:         "Page Title",
				Description:   "Description",
				BaseURL:       Must(url.Parse("https://events.testing")),
				SessionSecret: []byte("sess_secret"),
			})).To(Succeed())
		})

		Specify("insert is ignored", func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: false,
				Title:         "Page Title 2",
				Description:   "Description 2",
				BaseURL:       Must(url.Parse("https://events2.testing")),
				SessionSecret: []byte("sess_secret2"),
			})).To(Succeed())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings).To(PointTo(MatchAllFields(Fields{
				"IsInitialized": BeTrue(),
				"Title":         Equal("Page Title"),
				"Description":   Equal("Description"),
				"BaseURL":       HaveField("String()", "https://events.testing"),
				"SessionSecret": Equal([]byte("sess_secret")),
			})))
		})
	})
})

var _ = Describe("updating settings", func() {
	When("settings don't exist", func() {
		Specify("precondition failed error is returned", func(ctx SpecContext) {
			Expect(model.UpdateSettings(ctx, db, &domain.Settings{
				IsInitialized: true,
				Title:         "Page Title",
				Description:   "Description",
				BaseURL:       Must(url.Parse("https://events.testing")),
				SessionSecret: []byte("sess_secret"),
			})).To(MatchError(wreck.PreconditionFailed))
		})
	})

	When("settings exist", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: false,
				Title:         "Page Title",
				Description:   "Description",
				BaseURL:       Must(url.Parse("https://events.testing")),
				SessionSecret: []byte("sess_secret"),
			})).To(Succeed())
		})

		Specify("settings are updated", func(ctx SpecContext) {
			Expect(model.UpdateSettings(ctx, db, &domain.Settings{
				IsInitialized: true,
				Title:         "Page Title 2",
				Description:   "Description 2",
				BaseURL:       Must(url.Parse("https://events2.testing")),
				SessionSecret: []byte("sess_secret2"),
			})).To(Succeed())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings).To(PointTo(MatchAllFields(Fields{
				"IsInitialized": BeTrue(),
				"Title":         Equal("Page Title 2"),
				"Description":   Equal("Description 2"),
				"BaseURL":       HaveField("String()", "https://events2.testing"),
				"SessionSecret": Equal([]byte("sess_secret2")),
			})))
		})
	})
})
