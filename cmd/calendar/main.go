package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alexedwards/scs/bunstore"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/handler"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/mgnsk/calendar/server"
	"github.com/ringsaturn/tzf"
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
	cfg, err := LoadConfig()
	if err != nil {
		return calendar.Internal.New("error loading configuration", err)
	}

	databaseDir, err := filepath.Abs(cfg.DatabaseDir)
	if err != nil {
		return calendar.Internal.New("invalid database dir", err)
	}

	if err := os.MkdirAll(databaseDir, 0755); err != nil {
		return calendar.Internal.New("error creating database dir", err)
	}

	filename := filepath.Join(databaseDir, "calendar.sqlite")

	db := sqlite.NewDB(filename).Connect()
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("error closing database connection", slog.String("error", err.Error()))
		}
	}()

	if err := calendar.MigrateUp(db.DB); err != nil {
		return calendar.Internal.New("error migrating database", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case <-quit:
			cancel()
		}
	}()

	// Run SQL optimizer periodic task.
	g.Go(func() error {
		if err := sqlite.RunOptimizer(ctx, db.DB); err != nil {
			return calendar.Internal.New("error running sqlite optimizer", err)
		}
		return nil
	})

	// Run expired invites cleanup periodic task.
	g.Go(func() error {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil

			case <-ticker.C:
				if err := model.DeleteExpiredInvites(ctx, db); err != nil {
					return err
				}
			}
		}
	})

	// Initialize the session store.
	store, err := bunstore.New(db)
	if err != nil {
		return calendar.Internal.New("error creating sqlite session store", err)
	}

	sm := server.NewSessionManager(store)

	finder, err := tzf.NewDefaultFinder()
	if err != nil {
		return calendar.Internal.New("error creating tzf", err)
	}

	httpHandler := server.NewHandler(
		// Static assets.
		calendar.RegisterAssetsHandler,
		// Setup.
		func(mux *http.ServeMux) { handler.NewSetupHandler(db, sm).Register(mux) },
		// Settings.
		func(mux *http.ServeMux) { handler.NewSettingsHandler(db, sm).Register(mux) },
		// Authentication.
		func(mux *http.ServeMux) { handler.NewAuthenticationHandler(db, sm).Register(mux) },
		// Events.
		func(mux *http.ServeMux) { handler.NewEventsHandler(db, sm).Register(mux) },
		// Events management.
		func(mux *http.ServeMux) { handler.NewEditEventHandler(db, sm, finder).Register(mux) },
		// Users management.
		func(mux *http.ServeMux) { handler.NewUsersHandler(db, sm).Register(mux) },
		// Feeds.
		// TODO: proper caching middleware for RSS and calendar feeds.
		// Should support conditional get.
		func(mux *http.ServeMux) { handler.NewFeedHandler(db).Register(mux) },
	)

	s := http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      httpHandler,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelDebug),
	}

	g.Go(func() error {
		slog.Info(fmt.Sprintf("listening at %s", cfg.ListenAddr))

		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return calendar.Internal.New("error running server", err)
		}

		return nil
	})

	g.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		slog.Info("shutting down the server")

		if err := s.Shutdown(ctx); err != nil {
			return calendar.Internal.New("error shutting down server", err)
		}

		return nil
	})

	return g.Wait()
}
