package api

import "net/url"

// FeedConfig is common feed configuration.
type FeedConfig struct {
	Title   string
	BaseURL *url.URL
}
