package domain

import (
	"crypto/rand"
	"net/url"
)

// Settings is the settings domain model.
type Settings struct {
	IsInitialized bool
	Title         string
	Description   string
	BaseURL       *url.URL
	SessionSecret []byte
}

// NewDefaultSettings creates new default settings.
func NewDefaultSettings() *Settings {
	u, err := url.Parse("https://my-awesome-events.testing")
	if err != nil {
		panic(err)
	}

	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		panic(err)
	}

	return &Settings{
		IsInitialized: false,
		Title:         "My Awesome Events",
		Description:   "All the awesome events in one place!",
		BaseURL:       u,
		SessionSecret: b,
	}
}
