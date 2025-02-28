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
	DatabaseDir string
	CacheDir    string
	Host        string
	Development bool

	BaseURL *url.URL
}

// LoadConfig loads the configuration.
func LoadConfig() (*Config, error) {
	c := &Config{
		ListenAddr:  cmp.Or(os.Getenv("LISTEN_ADDR"), ":443"),
		DatabaseDir: os.Getenv("DATABASE_DIR"),
		CacheDir:    os.Getenv("CACHE_DIR"),
		Host:        os.Getenv("HOST"),
		Development: os.Getenv("MODE") == "development",
	}

	var errs []error

	if c.ListenAddr == "" {
		errs = append(errs, fmt.Errorf("listen_addr: is required"))
	}

	if c.DatabaseDir == "" {
		errs = append(errs, fmt.Errorf("database_dir: is required"))
	}

	if c.CacheDir == "" {
		errs = append(errs, fmt.Errorf("cache_dir: is required"))
	}

	if c.Host == "" {
		errs = append(errs, fmt.Errorf("host: is required"))
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", c.Host))
	if err != nil {
		errs = append(errs, fmt.Errorf("host: invalid host"))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	c.BaseURL = u

	return c, nil
}
