package server

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/calendar/pkg/wreck"
)

// Recover returns a middleware which recovers from panics anywhere in the chain
// and returns an error with stack trace.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (returnErr error) {
			defer func() {
				if r := recover(); r != nil {
					if r == http.ErrAbortHandler {
						panic(r)
					}

					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}

					var (
						stack  []byte
						length int
					)

					stack = make([]byte, 4<<10) // 4 KB
					length = runtime.Stack(stack, true)
					stack = stack[:length]

					returnErr = calendar.Internal.
						With(calendar.Stack, string(stack)).
						New("", err)
				}
			}()

			return next(c)
		}
	}
}

// Logger returns a logger with attributes populated from echo context.
func Logger(c echo.Context) *slog.Logger {
	req := c.Request()
	res := c.Response()

	return slog.With(
		"status", res.Status,
		"method", req.Method,
		"uri", req.RequestURI,
		"real_ip", c.RealIP(),
	)
}

// ErrorHandler is Echo server's error handler.
// It renders HTML error pages and logs errors.
func ErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		var (
			code = http.StatusInternalServerError
			msg  = "Something went wrong"
		)

		if errors.Is(err, context.DeadlineExceeded) {
			code = http.StatusGatewayTimeout
			msg = "Timeout"
		} else if werr := *new(wreck.Error); errors.As(err, &werr) {
			if v, ok := wreck.Value(werr, calendar.KeyHTTPCode); ok {
				code = int(v.Int64())
			}
			msg = cmp.Or(werr.Message(), msg)
		} else if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = cmp.Or(fmt.Sprint(he.Message), msg)
		}

		errText := fmt.Sprintf("Error %d: %s", code, msg)

		c.Response().Status = code

		if err := html.Page(html.PageProps{
			Title:        "Error",
			User:         nil,
			Path:         c.Path(),
			CSRF:         "",
			Children:     html.ErrorMain(errText),
			FlashSuccess: "",
		}).Render(c.Response()); err != nil {
			Logger(c).With(wreck.Args(err)...).Error("error rendering error page", slog.Any("reason", err))
		}

		logger := Logger(c).With(wreck.Args(err)...)

		switch {
		case c.Response().Status >= 500:
			logger.Error("server error", slog.Any("reason", err))

		case c.Response().Status >= 400 && c.Response().Status <= 403:
			logger.Error("client error", slog.Any("reason", err))
		}
	}
}
