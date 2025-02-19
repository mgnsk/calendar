package domain_test

import (
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/pkg/snowflake"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("setting user password", func() {
	var u *domain.User

	JustBeforeEach(func() {
		u = &domain.User{
			ID:       snowflake.Generate(),
			Username: "username",
		}
	})

	Specify("user password is hashed", func() {
		Expect(u.SetPassword("password")).To(Succeed())

		Expect(u.Password).NotTo(BeEmpty())
	})
})

var _ = Describe("verifying user password", func() {
	var u *domain.User

	JustBeforeEach(func() {
		u = &domain.User{
			ID:       snowflake.Generate(),
			Username: "username",
		}

		Expect(u.SetPassword("password")).To(Succeed())
	})

	Specify("user password can be verified", func() {
		Expect(u.VerifyPassword("password")).To(Succeed())
	})
})
