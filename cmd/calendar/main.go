package main

import (
	"context"
	"errors"
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
	session "github.com/canidam/echo-scs-session"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/handler"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/mgnsk/calendar/pkg/wreck"
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
		return wreck.Internal.New("error loading configuration", err)
	}

	databaseDir, err := filepath.Abs(cfg.DatabaseDir)
	if err != nil {
		return wreck.Internal.New("invalid database dir", err)
	}

	if err := os.MkdirAll(databaseDir, 0755); err != nil {
		return wreck.Internal.New("error creating database dir", err)
	}

	filename := filepath.Join(databaseDir, "calendar.sqlite")

	db := sqlite.NewDB(filename).Connect()
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("error closing database connection", slog.String("error", err.Error()))
		}
	}()

	if err := calendar.MigrateUp(db.DB); err != nil {
		return wreck.Internal.New("error migrating database", err)
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
			return wreck.Internal.New("error running sqlite optimizer", err)
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

	e := server.NewServer()

	// Initialize the session store.
	store, err := bunstore.New(db)
	if err != nil {
		return wreck.Internal.New("error creating sqlite session store", err)
	}

	sm := server.NewSessionManager(store, e)

	finder, err := tzf.NewDefaultFinder()
	if err != nil {
		return wreck.Internal.New("error creating tzf", err)
	}

	sessionMiddleware := session.LoadAndSaveWithConfig(session.Config{
		Skipper: nil,
		ErrorHandler: func(err error, c echo.Context) {
			if errors.Is(err, context.DeadlineExceeded) {
				return
			}

			server.Logger(c).With(wreck.Args(err)...).
				Error("session error", slog.Any("reason", err))
		},
		SessionManager: sm,
	})

	// Static assets.
	calendar.RegisterAssetsHandler(e)

	// Setup.
	{
		g := e.Group("",
			sessionMiddleware,
		)

		h := handler.NewSetupHandler(db, sm)
		h.Register(g)
	}

	// Authentication.
	{
		g := e.Group("",
			sessionMiddleware,
		)

		h := handler.NewAuthenticationHandler(db, sm)
		h.Register(g)
	}

	// Events.
	{
		g := e.Group("",
			sessionMiddleware,
		)

		h := handler.NewEventsHandler(db, sm)
		h.Register(g)
	}

	// Events management.
	{
		g := e.Group("",
			sessionMiddleware,
		)

		h := handler.NewEditEventHandler(db, sm, finder)
		h.Register(g)
	}

	// Users management.
	{
		g := e.Group("",
			sessionMiddleware,
		)

		h := handler.NewUsersHandler(db, sm)
		h.Register(g)
	}

	// Feeds.
	{
		// TODO: proper caching middleware for RSS and calendar feeds.
		// Should support conditional get.
		g := e.Group("")

		h := handler.NewFeedHandler(db, cfg.BaseURL)
		h.Register(g)
	}

	g.Go(func() error {
		slog.Info(fmt.Sprintf("listening at %s", cfg.ListenAddr))

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

	return g.Wait()
}
