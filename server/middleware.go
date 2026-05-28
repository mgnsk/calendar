package server

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"slices"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/model"
	"github.com/mgnsk/wreck"
	sloghttp "github.com/samber/slog-http"
	"github.com/uptrace/bun"
)

// MiddlewareFunc is a HTTP middleware function.
type MiddlewareFunc func(next http.Handler) http.Handler

// WithMiddleware applies middleware to run in order before handler.
func WithMiddleware(h http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for _, mw := range slices.Backward(middlewares) {
		h = mw(h)
	}
	return h
}

// NewTimeoutMiddleware creates a new context timeout middleware.
func NewTimeoutMiddleware(timeout time.Duration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ErrorHandler is a HTTP error handler middleware.
// It renders HTML error pages and logs errors.
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			var err error

			if r := recover(); r == nil {
				return
			} else if r == http.ErrAbortHandler {
				panic(r)
			} else if rerr, ok := r.(error); ok {
				err = rerr
			} else {
				err = fmt.Errorf("%v", r)
			}

			// Default status code and message.
			var (
				code = http.StatusInternalServerError
				msg  = "Something went wrong"
			)

			// Attempt to detect status code and message from error.
			if werr, ok := errors.AsType[*wreck.Error](err); ok {
				if v, ok := wreck.Value[int](werr, calendar.KeyHTTPCode); ok {
					code = v
				}
				msg = cmp.Or(werr.Message(), msg)
			} else if errors.Is(err, context.DeadlineExceeded) { // TODO: context.Canceled
				code = http.StatusGatewayTimeout
				msg = "Timeout"
			}

			stack := make([]byte, 4<<10) // 4 KB
			length := runtime.Stack(stack, true)
			stack = stack[:length]

			sloghttp.AddCustomAttributes(r, slog.String("err", err.Error()))
			sloghttp.AddCustomAttributes(r, slog.String("stack", string(stack)))

			w.WriteHeader(code)

			if err := html.Page(html.PageProps{
				Title:        "Error",
				User:         nil,
				Path:         r.URL.Path,
				Children:     html.ErrorMain(fmt.Sprintf("Error %d: %s", code, msg)),
				FlashSuccess: "",
			}).Render(w); err != nil {
				// TODO
				_ = err
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// NewSessionMiddleware creates a new session middleware.
func NewSessionMiddleware(sm *scs.SessionManager) MiddlewareFunc {
	return sm.LoadAndSave
}

// NewSettingsMiddleware creates a new settings middleware.
func NewSettingsMiddleware(db *bun.DB) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			settings, err := model.GetSettings(r.Context(), db)
			if err != nil {
				if !errors.Is(err, calendar.NotFound) {
					panic(err)
				}
			}

			if settings == nil && r.URL.Path != "/setup" {
				// First setup.
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), settingsCtxKey{}, settings))

			next.ServeHTTP(w, r)
		})
	}
}

// NewUserMiddleware creates a new user middleware.
func NewUserMiddleware(db *bun.DB, sm *scs.SessionManager) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if sm == nil {
				// Public endpoint.
				next.ServeHTTP(w, r)
				return
			}

			if username := sm.GetString(r.Context(), "username"); username != "" {
				user, err := model.GetUserByUsername(r.Context(), db, username)
				if err != nil {
					if !errors.Is(err, calendar.NotFound) {
						panic(err)
					}
				}

				if user == nil {
					// User has been deleted.
					if err := sm.Destroy(r.Context()); err != nil {
						panic(err)
					}
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}

				r = r.WithContext(context.WithValue(r.Context(), userCtxKey{}, user))
			}

			next.ServeHTTP(w, r)
		})
	}
}
