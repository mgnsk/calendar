package model_test

import (
	"time"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Note: basic insert/delete flows are covered by the invite and register HTTP
// integration tests (handler/users_test.go). Expired-invite cleanup remains
// because it is only ever invoked from the periodic background ticker in
// cmd/calendar/main.go, never reachable via HTTP.
var _ = Describe("deleting expired invites", func() {
	When("both expired and active invites exist", func() {
		var tokenFuture, tokenPast uuid.UUID

		BeforeEach(func(ctx SpecContext) {
			tokenFuture = uuid.New()
			tokenPast = uuid.New()

			By("inserting invite in future", func() {
				Expect(model.InsertInvite(ctx, db, &domain.Invite{
					Token:      tokenFuture,
					ValidUntil: time.Now().Add(time.Hour),
					CreatedBy:  snowflake.Generate(),
				})).To(Succeed())
			})

			By("inserting expired invite", func() {
				Expect(model.InsertInvite(ctx, db, &domain.Invite{
					Token:      tokenPast,
					ValidUntil: time.Now().Add(-time.Hour),
					CreatedBy:  snowflake.Generate(),
				})).To(Succeed())
			})
		})

		Specify("expired invites can be deleted", func(ctx SpecContext) {
			Expect(model.DeleteExpiredInvites(ctx, db)).To(Succeed())

			By("asserting invite in future exists", func() {
				invite := Must(model.GetInvite(ctx, db, tokenFuture))
				Expect(invite.Token).To(Equal(tokenFuture))
			})

			By("asserting expired invite was deleted", func() {
				Expect(model.GetInvite(ctx, db, tokenPast)).Error().To(MatchError(calendar.NotFound))
			})
		})
	})

	When("only active invites exist", func() {
		var tokenFuture uuid.UUID

		BeforeEach(func(ctx SpecContext) {
			tokenFuture = uuid.New()

			By("inserting invite in future", func() {
				Expect(model.InsertInvite(ctx, db, &domain.Invite{
					Token:      tokenFuture,
					ValidUntil: time.Now().Add(time.Hour),
					CreatedBy:  snowflake.Generate(),
				})).To(Succeed())
			})
		})

		Specify("deleting expired events is no-op", func(ctx SpecContext) {
			Expect(model.DeleteExpiredInvites(ctx, db)).To(Succeed())

			By("asserting invite in future exists", func() {
				invite := Must(model.GetInvite(ctx, db, tokenFuture))
				Expect(invite.Token).To(Equal(tokenFuture))
			})
		})
	})
})
