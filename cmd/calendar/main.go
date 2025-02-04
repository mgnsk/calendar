package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"math/rand/v2"
	"net/http"
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
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	slogecho "github.com/samber/slog-echo"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
)

func main() {
	log.SetFlags(0) // no time prefix

	if err := run(); err != nil {
		slog.Error("error running application", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	isDemo := flag.Bool("demo", false, "enables demo mode")
	flag.Parse()

	if *isDemo {
		slog.Info("running in demo mode")
	}

	cfg, err := LoadConfig()
	if err != nil {
		return wreck.Internal.New("error loading configuration", err)
	}

	databaseDir, err := filepath.Abs(cfg.DatabaseDir)
	if err != nil {
		return wreck.Internal.New("invalid database dir", err)
	}

	if err := os.MkdirAll(databaseDir, 0755); err != nil {
		return wreck.Internal.New("error creating database dir", err)
	}

	dsn := filepath.Join(databaseDir, "calendar.sqlite")
	db := sqlite.NewDB(dsn).Connect()
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("error closing database connection", slog.String("error", err.Error()))
		}
	}()

	if err := internal.MigrateUp(db.DB); err != nil {
		return wreck.Internal.New("error migrating database", err)
	}

	model.RegisterModels(db)

	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := model.InsertOrIgnoreSettings(ctx, db, domain.NewDefaultSettings()); err != nil {
			return wreck.Internal.New("unable to initialize settings", err)
		}
	}

	if *isDemo {
		go func() {
			n := 1000
			for range n {
				insertTestData(db)
			}
		}()
	}

	if err := ensureAdminExists(db); err != nil {
		return wreck.Internal.New("unable to create admin user", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	// g.Go(func() error {
	// 	if err := sqlite.RunOptimizer(ctx, db.DB); err != nil {
	// 		return wreck.Internal.New("error running sqlite optimizer", err)
	// 	}
	// 	return nil
	// })

	apiConfig := api.Config{
		PageTitle:     cfg.PageTitle,
		BaseURL:       cfg.BaseURL,
		SessionSecret: []byte(cfg.SessionSecret),
	}

	e := echo.New()
	e.Use(
		slogecho.NewWithConfig(slog.Default(), slogecho.Config{
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,

			WithUserAgent: true,
			WithRequestID: true,
		}),
		middleware.Recover(), // Recover from all panics to always have your server up.
		api.ErrorHandler(apiConfig),
		api.TimeoutMiddleware(time.Minute),
		// api.LoadSettingsMiddleware(db),
	)

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

	g.Go(func() error {
		if err := e.Start(cfg.ListenAddr); err != nil && err != http.ErrServerClosed {
			return wreck.Internal.New("error running server", err)
		}
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		slog.Info("shutting down the server")

		if err := e.Shutdown(ctx); err != nil {
			return wreck.Internal.New("error shutting down server", err)
		}

		return nil
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	errs := make(chan error, 1)
	go func() {
		errs <- g.Wait()
	}()

	select {
	case <-quit:
		cancel()
		err := <-errs
		return err

	case err := <-errs:
		cancel()
		return err
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

	tags1 := []string{"tag1", "tag2", "tag3", "tag4", "tag5", "tag6", "tag7"}
	tags2 := []string{"tag1", "tag2", "tag8", "tag9", "tag10"}
	tags3 := []string{"tag3", "tag11", "tag12", "tag13", "Some Tag"}

	ts := getRandBaseTime().Truncate(15 * time.Minute)
	n := rand.IntN(10000)
	event1 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(ts.Add(-48 * time.Hour)),
		EndAt:       timestamp.Timestamp{},
		Title:       fmt.Sprintf("Event %d", n),
		Description: "Desc 1",
		URL:         "https://event1.testing",
		Tags:        tags1,
	}

	ts = getRandBaseTime().Truncate(15 * time.Minute)
	event2 := &domain.Event{
		ID:      snowflake.Generate(),
		StartAt: timestamp.New(ts.Add(-12 * time.Hour)),
		EndAt:   timestamp.Timestamp{},
		Title:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec placerat nec enim sed pretium.",
		Description: `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec placerat nec enim sed pretium. Donec volutpat ornare convallis. Praesent cursus elementum felis, vel condimentum urna. Nullam feugiat, nunc eget vehicula aliquam, nunc neque molestie nunc, in rhoncus turpis ex blandit ante. Quisque rhoncus diam id vulputate suscipit. Etiam venenatis bibendum turpis mollis suscipit. Pellentesque nec tortor non nisi mollis euismod. Praesent felis lectus, eleifend nec orci in, fringilla tempor tortor.

Ut consectetur nulla quam, a tristique nibh volutpat quis. Praesent consequat mi nec orci suscipit ullamcorper. Vestibulum vitae eleifend justo. Nulla sed bibendum elit. Proin ultrices justo nec massa commodo, ut fringilla eros eleifend. Sed ligula diam, auctor sit amet tempor sit amet, commodo a quam. Sed et neque convallis, condimentum velit vel, interdum diam. Nunc at purus eget augue elementum viverra. Donec interdum lectus libero, sed gravida urna venenatis at. Praesent odio nibh, facilisis eu massa et, bibendum iaculis elit. Vivamus faucibus, turpis eget molestie consectetur, elit dui condimentum magna, sit amet congue nibh odio sed sapien. Maecenas vel dictum justo. Cras malesuada congue velit, sagittis convallis leo interdum ut. Proin fermentum dolor vel lacinia egestas.

Donec consectetur, erat vel egestas fringilla, justo leo tincidunt enim, at finibus arcu neque eu nunc. Ut consectetur semper nulla id elementum. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Curabitur laoreet lorem nec magna tempor venenatis. Vestibulum gravida in velit in mollis. Ut sodales tempus lectus sed malesuada. Nullam lacinia lacus non neque vehicula, et suscipit nunc dignissim. Aliquam et augue at lectus pellentesque suscipit eu a arcu. Nam vitae justo eros. Donec lacinia posuere molestie. Morbi id eros efficitur, dictum odio eget, congue lacus. Ut vel erat eu nisi iaculis tincidunt. Sed et ante ornare, vulputate massa et, posuere nibh. Integer scelerisque interdum tristique. Ut dapibus, elit sed imperdiet malesuada, eros augue sagittis nisi, at ultrices lacus neque ac nunc. In accumsan nec orci ut maximus.
		`,
		URL:  "https://event2.testing",
		Tags: tags2,
	}

	ts = getRandBaseTime().Truncate(15 * time.Minute)
	n = rand.IntN(10000)
	event3 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     timestamp.New(ts),
		EndAt:       timestamp.New(ts.Add(2 * time.Hour)),
		Title:       fmt.Sprintf("Event %d", n),
		Description: "Desc 3",
		URL:         "https://event3.testing",
		Tags:        tags3,
	}

	for _, ev := range []*domain.Event{event1, event2, event3} {
		if err := model.InsertEvent(context.Background(), db, ev); err != nil {
			panic(err)
		}
	}
}
