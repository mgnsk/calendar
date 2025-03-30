package model_test

import (
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	"github.com/mgnsk/calendar/pkg/wreck"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("inserting users", func() {
	When("user does not exist", func() {
		It("is inserted", func(ctx SpecContext) {
			Expect(model.InsertUser(ctx, db, &domain.User{
				ID:       snowflake.Generate(),
				Username: "username",
				Password: []byte("password"),
				Role:     domain.Admin,
			})).To(Succeed())

			user := Must(model.GetUser(ctx, db, "username"))

			Expect(user).To(PointTo(MatchAllFields(Fields{
				"ID":       Not(BeZero()),
				"Username": Equal("username"),
				"Password": Equal([]byte("password")),
				"Role":     Equal(domain.Admin),
			})))
		})
	})

	When("user exists", func() {
		JustBeforeEach(func(ctx SpecContext) {
			Expect(model.InsertUser(ctx, db, &domain.User{
				ID:       snowflake.Generate(),
				Username: "username",
				Password: []byte("password"),
				Role:     domain.Admin,
			})).To(Succeed())
		})

		Specify("already exists error is returned", func(ctx SpecContext) {
			err := model.InsertUser(ctx, db, &domain.User{
				ID:       snowflake.Generate(),
				Username: "username",
				Password: []byte("password"),
				Role:     domain.Admin,
			})

			Expect(err).To(MatchError(wreck.AlreadyExists))
		})
	})
})

var _ = Describe("deleting users", func() {
	JustBeforeEach(func(ctx SpecContext) {
		Expect(model.InsertUser(ctx, db, &domain.User{
			ID:       snowflake.Generate(),
			Username: "username",
			Password: []byte("password"),
			Role:     domain.Admin,
		})).To(Succeed())
	})

	Specify("user is deleted", func(ctx SpecContext) {
		users := Must(model.ListUsers(ctx, db))
		Expect(users).To(HaveLen(1))

		Expect(model.DeleteUser(ctx, db, "username")).To(Succeed())

		users = Must(model.ListUsers(ctx, db))
		Expect(users).To(HaveLen(0))
	})
})

var _ = Describe("listing users", func() {
	JustBeforeEach(func(ctx SpecContext) {
		for _, username := range []string{"user1", "user2"} {
			Expect(model.InsertUser(ctx, db, &domain.User{
				ID:       snowflake.Generate(),
				Username: username,
				Password: []byte("password"),
				Role:     domain.Admin,
			})).To(Succeed())
		}
	})

	Specify("users are listed in creation time asc", func(ctx SpecContext) {
		users := Must(model.ListUsers(ctx, db))

		Expect(users).To(HaveExactElements(
			HaveField("Username", "user1"),
			HaveField("Username", "user2"),
		))
	})
})
