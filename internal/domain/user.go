package domain

import (
	"errors"
	"time"

	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	"golang.org/x/crypto/bcrypt"
)

// Role is a user role.
type Role int

func (r Role) String() string {
	switch r {
	case Author:
		return "author"

	case Editor:
		return "editor"

	case Admin:
		return "admin"

	default:
		return "unknown"
	}
}

// User roles.
const (
	// Author can add events and edit their own events.
	Author Role = 0

	// Editor can add events and edit all events.
	Editor Role = 1

	// Admin can do everything.
	Admin Role = 255
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
			return wreck.InvalidValue.New("Password too long", err)
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
			return wreck.InvalidValue.New("Invalid credentials", err)
		}
		return err
	}

	return nil
}
