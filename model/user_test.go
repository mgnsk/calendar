package model_test

import (
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/mgnsk/calendar/pkg/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

// Note: basic insert / duplicate-username / delete flows are covered by the
// setup, register, and delete-user HTTP integration tests
// (handler/setup_test.go, handler/users_test.go). This case remains because
// no HTTP route exposes changing a user's own username/password together -
// POST /upgrade-user only ever mutates Role.
var _ = Describe("updating users", func() {
	When("user exists", func() {
		var userID snowflake.ID

		JustBeforeEach(func(ctx SpecContext) {
			userID = snowflake.Generate()

			Expect(model.InsertUser(ctx, db, &domain.User{
				ID:       userID,
				Username: "username",
				Password: []byte("password"),
				Role:     domain.Admin,
			})).To(Succeed())
		})

		Specify("user is updated", func(ctx SpecContext) {
			Expect(model.UpdateUser(ctx, db, &domain.User{
				ID:       userID,
				Username: "username2",
				Password: []byte("password2"),
				Role:     domain.Author,
			})).To(Succeed())

			user := Must(model.GetUserByUsername(ctx, db, "username2"))
			Expect(user).To(PointTo(MatchAllFields(Fields{
				"ID":       Equal(userID),
				"Username": Equal("username2"),
				"Password": Equal([]byte("password2")),
				"Role":     Equal(domain.Author),
			})))
		})
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
