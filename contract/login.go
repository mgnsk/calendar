package contract

import "net/url"

// LoginForm is a login form.
type LoginForm struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

// Validate the form.
func (f *LoginForm) Validate() url.Values {
	errs := url.Values{}

	if f.Username == "" {
		errs.Set("username", "Required")
	}

	if f.Password == "" {
		errs.Set("password", "Required")
	}

	return errs
}
