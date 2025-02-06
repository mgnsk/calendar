package domain

// Settings is the settings domain model.
type Settings struct {
	IsInitialized bool
	Title         string
	Description   string
}

// NewDefaultSettings creates new default settings.
func NewDefaultSettings() *Settings {
	return &Settings{
		IsInitialized: false,
		Title:         "My Awesome Events",
		Description:   "All the awesome events in one place!",
	}
}
