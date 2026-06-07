package contract

import "net/url"

// SettingsForm is the settings form.
type SettingsForm struct {
	Title       string `in:"form=pagetitle"`
	Description string `in:"form=pagedesc"`
	Stopwords   string `in:"form=stopwords"`
}

// Validate the form.
func (f *SettingsForm) Validate() url.Values {
	errs := url.Values{}

	if f.Title == "" {
		errs.Set("pagetitle", "Title must be set")
	}

	return errs
}
