package contract

import (
	"net/url"

	"github.com/google/uuid"
	"github.com/mgnsk/calendar/pkg/snowflake"
)

// DeleteUserRequest is a request to delete a user.
type DeleteUserRequest struct {
	UserID snowflake.ID `form:"user_id"`
}

// RegisterRequest is a request to render the register page.
type RegisterRequest struct {
	Token uuid.UUID `param:"token"`
}

// RegisterForm is the register form.
type RegisterForm struct {
	Username  string `form:"username"`
	Password1 string `form:"password1"`
	Password2 string `form:"password2"`
}

// Validate the form.
func (f *RegisterForm) Validate() url.Values {
	errs := url.Values{}

	if f.Username == "" {
		errs.Set("username", "Username must be set")
	} else if len(f.Username) < 3 {
		errs.Set("username", "Username must be at least 3 characters")
	} else if len(f.Username) > 30 {
		errs.Set("username", "Username must be at most 30 characters")
	}

	// TODO: password strength check
	if f.Password1 == "" {
		errs.Set("password1", "Password must be set")
	}

	if f.Password2 == "" {
		errs.Set("password2", "Password must be set")
	}

	if f.Password1 != f.Password2 {
		errs.Set("password2", "Passwords must match")
	}

	return errs
}
