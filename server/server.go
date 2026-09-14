package server

import (
	"log/slog"
	"net/http"
	"time"

	httpin_integration "github.com/ggicci/httpin/integration"
	sloghttp "github.com/samber/slog-http"
)

func init() {
	httpin_integration.UseHttpPathVariable("path")
}

// NewHandler builds the application's HTTP handler. Each registrar is called
// with the mux to register its routes, then the result is wrapped with the
// global middleware.
func NewHandler(registrars ...func(*http.ServeMux)) http.Handler {
	mux := http.NewServeMux()

	for _, register := range registrars {
		register(mux)
	}

	return WithMiddleware(mux,
		sloghttp.NewWithConfig(slog.Default(), sloghttp.Config{
			DefaultLevel:     slog.LevelInfo,
			ClientErrorLevel: slog.LevelWarn,
			ServerErrorLevel: slog.LevelError,

			WithUserAgent:      true,
			WithRequestID:      true,
			WithRequestBody:    false,
			WithRequestHeader:  false,
			WithResponseBody:   false,
			WithResponseHeader: false,
			WithSpanID:         false,
			WithTraceID:        false,
			WithClientIP:       true,
			WithCustomMessage:  nil,

			Filters: []sloghttp.Filter{
				func(w sloghttp.WrapResponseWriter, _ *http.Request) bool {
					if w.Status() >= 500 {
						return true
					}

					if w.Status() >= 400 && w.Status() <= 403 {
						return true
					}

					return false
				},
			},
		}),
		ErrorHandler,
		NewTimeoutMiddleware(time.Minute),
	)
}
