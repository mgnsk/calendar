package main

import (
	"context"
	_ "embed"
	"log"
	"log/slog"
	"math/rand/v2"
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

	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	databaseDir, err := filepath.Abs(cfg.DatabaseDir)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(databaseDir, 0755); err != nil {
		log.Fatal(err)
	}

	dsn := filepath.Join(databaseDir, "calendar.sqlite")
	db := sqlite.NewDB(dsn).Connect()

	if err := internal.MigrateUp(db.DB); err != nil {
		log.Fatal(err)
	}

	model.RegisterModels(db)

	insertTestData(db)
	if err := ensureAdminExists(db); err != nil {
		log.Fatalf("unable to create admin user: %s", err.Error())
	}

	e := echo.New()
	e.Use(
		slogecho.New(slog.Default()), // Log everything.
		middleware.Recover(),         // Recover from all panics to always have your server up
	)

	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		log.Fatal(err)
	}

	apiConfig := api.Config{
		PageTitle:     cfg.PageTitle,
		BaseURL:       baseURL,
		SessionSecret: []byte(cfg.SessionSecret),
	}

	{
		h := api.NewFeedHandler(db, apiConfig)
		h.Register(e)
	}

	{
		h := api.NewHTMLHandler(db, apiConfig)
		h.Register(e)
	}

	e.Server.ReadHeaderTimeout = time.Minute
	e.Server.WriteTimeout = time.Minute

	// Start server
	go func() {
		if err := e.Start(cfg.ListenAddr); err != nil && err != http.ErrServerClosed {
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

// ensureAdminExists inserts an admin:admin user when no users are present in the database.
func ensureAdminExists(db *bun.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	users, err := model.ListUsers(ctx, db)
	if err != nil {
		return err
	}

	if len(users) > 0 {
		return nil
	}

	slog.Info("Inserting admin:admin user")

	user := &domain.User{
		ID:       snowflake.Generate(),
		Username: "admin",
		Role:     domain.Admin,
	}
	user.SetPassword("admin")

	return model.InsertUser(ctx, db, user)
}

func insertTestData(db *bun.DB) {
	getRandBaseTime := func() time.Time {
		baseTime := time.Now()

		var hours time.Duration
		if rand.Int()%2 == 0 {
			hours = rand.N(30 * 24 * time.Hour)
		}

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
		Tags:        []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7"},
	}

	event2 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(getRandBaseTime().Add(-12 * time.Hour)),
		EndAt:       timestamp.Timestamp{},
		Title:       "Event 2",
		Description: "Desc 2",
		URL:         "https://event2.testing",
		Tags:        []string{"tag1", "tag2", "tag8", "tag9", "tag10"},
	}

	ts := getRandBaseTime()
	event3 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(ts),
		EndAt:       timestamp.New(ts.Add(2 * time.Hour)),
		Title:       "Event 3",
		Description: "Desc 3",
		URL:         "https://event3.testing",
		Tags:        []string{"tag3", "tag11", "tag12", "tag13", "Some Tag"},
	}

	for _, ev := range []*domain.Event{event1, event2, event3} {
		if err := model.InsertEvent(context.Background(), db, ev); err != nil {
			panic(err)
		}
	}
}
