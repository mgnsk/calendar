package main

import (
	"context"
	"math/rand/v2"

	_ "embed"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/api"
	"github.com/mgnsk/calendar/internal/domain"
	"github.com/mgnsk/calendar/internal/model"
	"github.com/mgnsk/calendar/internal/pkg/snowflake"
	"github.com/mgnsk/calendar/internal/pkg/sqlite"
	"github.com/mgnsk/calendar/internal/pkg/timestamp"

	slogecho "github.com/samber/slog-echo"
	"github.com/uptrace/bun"
)

func main() {
	log.SetFlags(0) // no time prefix

	var (
		addr        string
		databaseDir string
	)

	flag.StringVar(&addr, "addr", ":8080", "listen address")
	flag.StringVar(&databaseDir, "database-dir", "", "database directory")
	flag.Parse()

	if addr == "" {
		log.Println("addr must not be empty")
		flag.Usage()
		os.Exit(1)
	}

	if databaseDir == "" {
		log.Println("database-dir must not be empty")
		flag.Usage()
		os.Exit(1)
	}

	dir, err := filepath.Abs(databaseDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}

	dsn := filepath.Join(dir, "calendar.sqlite")
	db := sqlite.NewDB(dsn).Connect()

	if err := internal.MigrateUp(db.DB); err != nil {
		log.Fatal(err)
	}

	model.RegisterModels(db)

	insertTestData(db)

	e := echo.New()
	e.Use(
		slogecho.New(slog.Default()), // Log everything.
		middleware.Recover(),         // Recover from all panics to always have your server up
		middleware.StaticWithConfig(middleware.StaticConfig{ // Serve assets from the embed filesystem.
			Filesystem: http.FS(internal.DistFS),
		}),
	)

	{
		// TODO
		baseURL, err := url.Parse("https://example.testing")
		if err != nil {
			log.Fatal(err)
		}

		// TODO
		feedConfig := api.FeedConfig{
			Title:   "My Feed",
			BaseURL: baseURL,
		}

		h := api.NewFeedHandler(db, feedConfig)
		h.Register(e)
	}

	{
		h := api.NewHTMLHandler(db)
		h.Register(e)
	}

	e.Server.ReadHeaderTimeout = time.Minute
	e.Server.WriteTimeout = time.Minute

	// Start server
	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Wait for exit signal.
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func insertTestData(db *bun.DB) {
	getRandBaseTime := func() time.Time {
		baseTime := time.Now()

		hours := rand.N(30 * 24 * time.Hour)

		baseTime = baseTime.Add(hours)

		return baseTime
	}

	event1 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(getRandBaseTime().Add(-48 * time.Hour)),
		EndAt:       timestamp.Timestamp{},
		Title:       "Event 1",
		Description: "Desc 1",
		URL:         "https://event1.testing",
		Tags:        []string{"tag1"},
	}

	event2 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(getRandBaseTime().Add(-12 * time.Hour)),
		EndAt:       timestamp.Timestamp{},
		Title:       "Event 2",
		Description: "Desc 2",
		URL:         "https://event2.testing",
		Tags:        []string{"tag1", "tag2"},
	}

	ts := getRandBaseTime()
	event3 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(ts),
		EndAt:       timestamp.New(ts.Add(2 * time.Hour)),
		Title:       "Event 3",
		Description: "Desc 3",
		URL:         "https://event3.testing",
		Tags:        []string{"tag3"},
	}

	for _, ev := range []*domain.Event{event1, event2, event3} {
		if err := model.InsertEvent(context.Background(), db, ev); err != nil {
			panic(err)
		}
	}
}
