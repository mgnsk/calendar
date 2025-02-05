package main

import (
	"cmp"
	"errors"
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

// Config is the calendar configuration.
type Config struct {
	ListenAddr  string
	DatabaseDir string
}

// LoadConfig loads the configuration.
func LoadConfig() (*Config, error) {
	c := &Config{
		ListenAddr:  cmp.Or(os.Getenv("LISTEN_ADDR"), ":8080"),
		DatabaseDir: os.Getenv("DATABASE_DIR"),
	}

	var errs []error

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
