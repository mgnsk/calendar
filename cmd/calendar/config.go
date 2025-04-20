package main

import (
	"cmp"
	"errors"
	"fmt"
	"net/url"
	"os"
)

// Config is the calendar configuration.
type Config struct {
	ListenAddr  string
	BaseURL     *url.URL
	DatabaseDir string
}

// LoadConfig loads the configuration.
func LoadConfig() (*Config, error) {
	var errs []error

	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		errs = append(errs, fmt.Errorf("hostname: is required"))
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", hostname))
	if err != nil {
		errs = append(errs, fmt.Errorf("hostname: error parsing base URL for hostname: %w", err))
	}

	c := &Config{
		ListenAddr:  cmp.Or(os.Getenv("LISTEN_ADDR"), ":8080"),
		DatabaseDir: os.Getenv("DATABASE_DIR"),
		BaseURL:     u,
	}

	if c.ListenAddr == "" {
		errs = append(errs, fmt.Errorf("listen_addr: is required"))
	}

	if c.DatabaseDir == "" {
		errs = append(errs, fmt.Errorf("database_dir: is required"))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return c, nil
}
