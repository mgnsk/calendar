package domain

// Settings is the settings domain model.
type Settings struct {
	Title       string
	Description string
}

// NewDefaultSettings creates new default settings.
func NewDefaultSettings() *Settings {
	return &Settings{
		Title:       "My Awesome Events",
		Description: "All the awesome events in one place!",
	}
}
