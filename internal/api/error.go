package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	slogecho "github.com/samber/slog-echo"
)

// ErrorHandler handles setting response status code from error and renders an error page.
func ErrorHandler(config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			var (
				msg  string
				code = http.StatusInternalServerError
			)

			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				code = http.StatusRequestTimeout
			} else if werr := *new(wreck.Error); errors.As(err, &werr) {
				msg = werr.Message()
				if v := wreck.Value(werr, wreck.KeyHTTPCode); v != nil {
					code = v.(int)
				}
			} else if he, ok := err.(*echo.HTTPError); ok {
				code = he.Code
			}

			slogecho.AddCustomAttributes(c, slog.String("error", err.Error()))
			c.Response().Status = code

			return html.ErrorPage(config.PageTitle, code, msg).Render(c.Response())
		}
	}
}
