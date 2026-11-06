package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/bunstore"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/handler"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/server"
	"github.com/mgnsk/webkit/app"
	"github.com/mgnsk/webkit/sqlite"
	"github.com/ringsaturn/tzf"
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

	if err := sqlite.Optimize(context.Background(), db.DB); err != nil {
		return calendar.Internal.New("error optimizing database", err)
	}

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

	s := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      httpHandler,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelDebug),
	}

	return app.Run(context.Background(), s,
		// Run SQL optimizer periodic task.
		app.Every(24*time.Hour, func(ctx context.Context, _ time.Time) error {
			return sqlite.Optimize(ctx, db.DB)
		}),
		// Run expired invites cleanup periodic task.
		app.Every(5*time.Minute, func(ctx context.Context, _ time.Time) error {
			return model.DeleteExpiredInvites(ctx, db)
		}),
	)
}
