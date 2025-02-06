package model_test

import (
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
			})).To(Succeed())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings).To(PointTo(MatchAllFields(Fields{
				"IsInitialized": BeFalse(),
				"Title":         Equal("Page Title"),
				"Description":   Equal("Description"),
			})))
		})
	})

	When("settings exist", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: true,
				Title:         "Page Title",
				Description:   "Description",
			})).To(Succeed())
		})

		Specify("insert is ignored", func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: false,
				Title:         "Page Title 2",
				Description:   "Description 2",
			})).To(Succeed())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings).To(PointTo(MatchAllFields(Fields{
				"IsInitialized": BeTrue(),
				"Title":         Equal("Page Title"),
				"Description":   Equal("Description"),
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
			})).To(MatchError(wreck.PreconditionFailed))
		})
	})

	When("settings exist", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertOrIgnoreSettings(ctx, db, &domain.Settings{
				IsInitialized: false,
				Title:         "Page Title",
				Description:   "Description",
			})).To(Succeed())
		})

		Specify("settings are updated", func(ctx SpecContext) {
			Expect(model.UpdateSettings(ctx, db, &domain.Settings{
				IsInitialized: true,
				Title:         "Page Title 2",
				Description:   "Description 2",
			})).To(Succeed())

			settings := Must(model.GetSettings(ctx, db))
			Expect(settings).To(PointTo(MatchAllFields(Fields{
				"IsInitialized": BeTrue(),
				"Title":         Equal("Page Title 2"),
				"Description":   Equal("Description 2"),
			})))
		})
	})
})
