package main

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// Config is the calendar configuration.
type Config struct {
	ListenAddr  string
	DatabaseDir string

	BaseURL       *url.URL
	PageTitle     string
	SessionSecret string
}

// LoadConfig loads the configuration.
func LoadConfig() (*Config, error) {
	c := &Config{
		ListenAddr:    cmp.Or(os.Getenv("LISTEN_ADDR"), ":8080"),
		DatabaseDir:   os.Getenv("DATABASE_DIR"),
		BaseURL:       nil,
		PageTitle:     cmp.Or(os.Getenv("PAGE_TITLE"), "My Awesome Calendar"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
	}

	baseURL := os.Getenv("BASE_URL")

	var errs []error

	if c.ListenAddr == "" {
		errs = append(errs, fmt.Errorf("listen_addr: is required"))
	}

	if c.DatabaseDir == "" {
		errs = append(errs, fmt.Errorf("database_dir: is required"))
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		errs = append(errs, fmt.Errorf("base_url: %w", err))
	}
	c.BaseURL = u

	if c.PageTitle == "" {
		errs = append(errs, fmt.Errorf("page_title: is required"))
	}

	if len(c.PageTitle) > 64 {
		errs = append(errs, fmt.Errorf("page_title: maximum len 64"))
	}

	if len(c.SessionSecret) != 32 {
		errs = append(errs, fmt.Errorf("session_secret: must be 32 bytes"))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return c, nil
}
