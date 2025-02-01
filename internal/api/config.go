package api

import "net/url"

// Config is common feed and web app configuration.
type Config struct {
	PageTitle     string
	BaseURL       *url.URL
	SessionSecret []byte
}
