package model_test

import (
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Note: the happy-path insert/update flows are covered by the setup and
// settings HTTP integration tests (handler/setup_test.go,
// handler/settings_test.go). This case remains because it is unreachable via
// HTTP - the settings-required middleware gate prevents ever calling
// UpdateSettings on a nonexistent row through a real request.
var _ = Describe("updating settings", func() {
	When("settings don't exist", func() {
		Specify("precondition failed error is returned", func(ctx SpecContext) {
			Expect(model.UpdateSettings(ctx, db, &domain.Settings{
				Title:       "Page Title",
				Description: "Description",
			})).To(MatchError(calendar.PreconditionFailed))
		})
	})
})
