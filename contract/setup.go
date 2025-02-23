package contract

import "net/url"

// SetupForm is a setup form.
type SetupForm struct {
	Title       string `form:"pagetitle"`
	Description string `form:"pagedesc"`
	Username    string `form:"username"`
	Password1   string `form:"password1"`
	Password2   string `form:"password2"`
}

// Validate the form.
func (f *SetupForm) Validate() url.Values {
	errs := url.Values{}

	if f.Title == "" {
		errs.Set("pagetitle", "Title must be set")
	}

	if f.Username == "" {
		errs.Set("username", "Username must be set")
	}

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
