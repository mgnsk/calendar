package model_test

import (
	"time"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	"github.com/mgnsk/calendar/pkg/wreck"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("inserting invites", func() {
	It("is inserted", func(ctx SpecContext) {
		token := uuid.New()
		createdBy := snowflake.Generate()

		Expect(model.InsertInvite(ctx, db, &domain.Invite{
			Token:      token,
			ValidUntil: time.Now(),
			CreatedBy:  createdBy,
		})).To(Succeed())

		invite := Must(model.GetInvite(ctx, db, token))

		Expect(invite.Token).To(Equal(token))
		Expect(invite.ValidUntil).To(BeTemporally("~", time.Now(), time.Second))
		Expect(invite.CreatedBy).To(Equal(createdBy))
	})
})

var _ = Describe("deleting invites", func() {
	var token uuid.UUID

	BeforeEach(func(ctx SpecContext) {
		token = uuid.New()

		Expect(model.InsertInvite(ctx, db, &domain.Invite{
			Token:      token,
			ValidUntil: time.Now(),
			CreatedBy:  snowflake.Generate(),
		})).To(Succeed())
	})

	Specify("invite can be deleted", func(ctx SpecContext) {
		Expect(model.DeleteInvite(ctx, db, token)).To(Succeed())

		_, err := model.GetInvite(ctx, db, token)
		Expect(err).To(MatchError(wreck.NotFound))
	})
})

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
				Expect(model.GetInvite(ctx, db, tokenPast)).Error().To(MatchError(wreck.NotFound))
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
