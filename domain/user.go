package domain

import (
	"errors"
	"time"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"golang.org/x/crypto/bcrypt"
)

// Role is a user role.
type Role string

// User roles.
const (
	// Author can add events and edit their own events.
	Author Role = "author"

	// Admin can do everything, including adding new users.
	Admin Role = "admin"
)

// User is the user domain model.
type User struct {
	ID       snowflake.ID
	Username string
	Password []byte
	Role     Role
}

// GetCreatedAt returns the user's created at time.
func (u *User) GetCreatedAt() time.Time {
	return snowflake.ParseTime(u.ID.Int64())
}

// SetPassword sets the user's password.
func (u *User) SetPassword(password string) error {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return calendar.InvalidValue.New("Password too long", err)
		}
		return err
	}

	u.Password = h

	return nil
}

// VerifyPassword verifies the user's password.
func (u *User) VerifyPassword(password string) error {
	if err := bcrypt.CompareHashAndPassword(u.Password, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return calendar.InvalidValue.New("Invalid credentials", err)
		}
		return err
	}

	return nil
}
