package main

import (
	"context"
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

	"github.com/alexedwards/scs/bunstore"
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	"github.com/mgnsk/calendar/handler"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/calendar/pkg/snowflake"
	"github.com/mgnsk/calendar/pkg/sqlite"
	"github.com/mgnsk/calendar/pkg/wreck"
	slogecho "github.com/samber/slog-echo"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/acme/autocert"
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

	g.Go(func() error {
		if err := sqlite.RunOptimizer(ctx, db.DB); err != nil {
			return wreck.Internal.New("error running sqlite optimizer", err)
		}
		return nil
	})

	// if *isDemo {
	// 	g.Go(func() error {
	// 		n := 1000
	// 		slog.Info(fmt.Sprintf("running in demo mode, inserting %d testdata", n))
	// 		for i := range n {
	// 			if err := insertTestData(ctx, db); err != nil {
	// 				return err
	// 			}
	// 			if i%1000 == 0 {
	// 				slog.Info(fmt.Sprintf("inserting testdata %d%% complete", int(float64(i)/float64(n)*100)))
	// 			}
	// 		}
	// 		slog.Info("finished inserting testdata")
	// 		return nil
	// 	})
	// }

	// Initialize the session store.
	store, err := bunstore.New(db)
	if err != nil {
		return wreck.Internal.New("error creating sqlite session store", err)
	}

	defer store.StopCleanup()

	// Initialize a new session manager and configure the session lifetime.
	sm := scs.New()
	sm.Store = store
	sm.HashTokenInStore = true
	sm.Lifetime = 24 * time.Hour
	// sm.IdleTimeout = 20 * time.Minute // TODO
	sm.Cookie.Name = "session_id"
	sm.Cookie.Domain = ""
	sm.Cookie.HttpOnly = true
	sm.Cookie.Path = "/"
	sm.Cookie.Persist = true
	sm.Cookie.SameSite = http.SameSiteStrictMode
	sm.Cookie.Secure = true

	e := echo.New()
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if err := handler.HandleError(err, c); err != nil {
			panic(err)
		}
	}
	sm.ErrorFunc = func(w http.ResponseWriter, r *http.Request, err error) {
		if err := handler.HandleError(err, e.NewContext(r, w)); err != nil {
			panic(err)
		}
	}

	e.Use(
		slogecho.NewWithConfig(slog.Default(), slogecho.Config{
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,

			WithUserAgent: true,
			WithRequestID: true,
		}),
		middleware.Recover(), // Recover from all panics to always have your server up.
		handler.ErrorHandler(),
		middleware.RequestID(),
		middleware.SecureWithConfig(middleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'",
			HSTSPreloadEnabled:    false,
		}),
		middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)),
		middleware.BodyLimit("1M"),
		middleware.ContextTimeout(time.Minute),
	)

	// Static assets.
	calendar.RegisterAssetsHandler(e)

	handler.Register(e, db, sm, cfg.BaseURL)

	e.Server.ReadHeaderTimeout = time.Minute
	e.Server.ReadTimeout = time.Minute
	e.Server.WriteTimeout = time.Minute
	e.Server.IdleTimeout = time.Minute

	g.Go(func() error {
		slog.Info(fmt.Sprintf("listening at %s", cfg.ListenAddr))

		if cfg.Development {
			if err := e.StartTLS(cfg.ListenAddr, "./certs/calendar.testing.crt", "./certs/calendar.testing.key"); err != nil && err != http.ErrServerClosed {
				return wreck.Internal.New("error running server", err)
			}
		} else {
			e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(cfg.DomainName)
			// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
			e.AutoTLSManager.Cache = autocert.DirCache(cfg.CacheDir)
			if err := e.StartAutoTLS(cfg.ListenAddr); err != nil && err != http.ErrServerClosed {
				return wreck.Internal.New("error running server", err)
			}
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

func insertTestData(ctx context.Context, db *bun.DB) error {
	getRandBaseTime := func() time.Time {
		baseTime := time.Now()

		var hours time.Duration
		if rand.Int()%2 == 0 {
			hours = rand.N(30 * 24 * time.Hour)
		}
		if rand.Int()%2 == 0 {
			hours *= -1
		}

		baseTime = baseTime.Add(hours)

		return baseTime
	}

	ts := getRandBaseTime().Truncate(15 * time.Minute)
	n := rand.IntN(10000)
	event1 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     ts.Add(-48 * time.Hour),
		Title:       fmt.Sprintf("Event %d", n),
		Description: "Desc 1",
		URL:         "https://event1.testing",
	}

	ts = getRandBaseTime().Truncate(15 * time.Minute)
	event2 := &domain.Event{
		ID:      snowflake.Generate(),
		StartAt: ts.Add(-12 * time.Hour),
		Title:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec placerat nec enim sed pretium.",
		Description: `
😀😀😀

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Donec placerat nec enim sed pretium. Donec volutpat ornare convallis. Praesent cursus elementum felis, vel condimentum urna. Nullam feugiat, nunc eget vehicula aliquam, nunc neque molestie nunc, in rhoncus turpis ex blandit ante. Quisque rhoncus diam id vulputate suscipit. Etiam venenatis bibendum turpis mollis suscipit. Pellentesque nec tortor non nisi mollis euismod. Praesent felis lectus, eleifend nec orci in, fringilla tempor tortor.

<b>HTML test</b>
😀😀😀

Ut consectetur nulla quam, a tristique nibh volutpat quis. Praesent consequat mi nec orci suscipit ullamcorper. Vestibulum vitae eleifend justo. Nulla sed bibendum elit. Proin ultrices justo nec massa commodo, ut fringilla eros eleifend. Sed ligula diam, auctor sit amet tempor sit amet, commodo a quam. Sed et neque convallis, condimentum velit vel, interdum diam. Nunc at purus eget augue elementum viverra. Donec interdum lectus libero, sed gravida urna venenatis at. Praesent odio nibh, facilisis eu massa et, bibendum iaculis elit. Vivamus faucibus, turpis eget molestie consectetur, elit dui condimentum magna, sit amet congue nibh odio sed sapien. Maecenas vel dictum justo. Cras malesuada congue velit, sagittis convallis leo interdum ut. Proin fermentum dolor vel lacinia egestas.

Donec consectetur, erat vel egestas fringilla, justo leo tincidunt enim, at finibus arcu neque eu nunc. Ut consectetur semper nulla id elementum. Orci varius natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Curabitur laoreet lorem nec magna tempor venenatis. Vestibulum gravida in velit in mollis. Ut sodales tempus lectus sed malesuada. Nullam lacinia lacus non neque vehicula, et suscipit nunc dignissim. Aliquam et augue at lectus pellentesque suscipit eu a arcu. Nam vitae justo eros. Donec lacinia posuere molestie. Morbi id eros efficitur, dictum odio eget, congue lacus. Ut vel erat eu nisi iaculis tincidunt. Sed et ante ornare, vulputate massa et, posuere nibh. Integer scelerisque interdum tristique. Ut dapibus, elit sed imperdiet malesuada, eros augue sagittis nisi, at ultrices lacus neque ac nunc. In accumsan nec orci ut maximus.
		`,
		URL: "https://event2.testing",
	}

	ts = getRandBaseTime().Truncate(15 * time.Minute)
	n = rand.IntN(10000)
	event3 := &domain.Event{
		ID:          snowflake.Generate(),
		StartAt:     ts,
		Title:       fmt.Sprintf("Event %d", n),
		Description: "Desc 3",
		URL:         "https://event3.testing",
	}

	for _, ev := range []*domain.Event{event1, event2, event3} {
		if err := model.InsertEvent(ctx, db, ev); err != nil {
			return err
		}
	}

	return nil
}
