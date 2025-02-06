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
	DomainName  string
	Development bool

	BaseURL *url.URL
}

// LoadConfig loads the configuration.
func LoadConfig() (*Config, error) {
	c := &Config{
		ListenAddr:  cmp.Or(os.Getenv("LISTEN_ADDR"), ":443"),
		DatabaseDir: os.Getenv("DATABASE_DIR"),
		CacheDir:    os.Getenv("CACHE_DIR"),
		DomainName:  os.Getenv("DOMAIN_NAME"),
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

	if c.DomainName == "" {
		errs = append(errs, fmt.Errorf("domain_name: is required"))
	}

	u, err := url.Parse(fmt.Sprintf("https://%s", c.DomainName))
	if err != nil {
		errs = append(errs, fmt.Errorf("domain_name: invalid domain name"))
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	c.BaseURL = u

	return c, nil
}
