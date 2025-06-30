package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
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

					returnErr = wreck.Internal.
						With(wreck.Stack, string(stack)).
						New("recovered panic", err)
				}
			}()

			return next(c)
		}
	}
}

// HandleError is a custom function to handle errors.
func HandleError(err error, c echo.Context) error {
	if c.Response().Committed {
		return nil
	}

	var (
		msg  string
		code = http.StatusInternalServerError

		logAttrs []any
	)

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		code = http.StatusRequestTimeout
	} else if werr := *new(wreck.Error); errors.As(err, &werr) {
		msg = werr.Message()

		if v, ok := wreck.Value(werr, wreck.KeyHTTPCode); ok {
			code = int(v.Int64())
		}

		logAttrs = append(logAttrs, wreck.Args(err)...)
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	reqID := c.Response().Header().Get(echo.HeaderXRequestID)
	c.Response().Status = code

	req := c.Request()
	res := c.Response()

	logger := slog.With(
		"reason", err,
		"status", code,
		"method", req.Method,
		"uri", req.RequestURI,
		"request_id", reqID,
		"real_ip", c.RealIP(),
	)

	if len(logAttrs) > 0 {
		logger = logger.With(logAttrs...)
	}

	switch {
	case res.Status >= 500:
		logger.ErrorContext(c.Request().Context(), "server error")

	case res.Status >= 400 && res.Status <= 403:
		logger.ErrorContext(c.Request().Context(), "client error")
	}

	errText := fmt.Sprintf("Error %d: %s (request ID: %s)", code, msg, reqID)

	return html.Page(html.PageProps{
		Title:        "Error",
		User:         nil,
		Path:         c.Path(),
		CSRF:         "",
		Children:     html.ErrorMain(errText),
		FlashSuccess: "",
	}).Render(c.Response())
}
