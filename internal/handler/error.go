package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/internal/html"
	"github.com/mgnsk/calendar/internal/pkg/wreck"
	slogecho "github.com/samber/slog-echo"
)

// ErrorHandler handles setting response status code from error and renders an error page.
func ErrorHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err == nil {
				return nil
			}

			return HandleError(err, c)
		}
	}
}

// HandleError is a custom function to handle errors.
func HandleError(err error, c echo.Context) error {
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

	reqID := c.Response().Header().Get(echo.HeaderXRequestID)

	slogecho.AddCustomAttributes(c, slog.String("error", err.Error()))
	c.Response().Status = code

	user := loadUser(c)
	csrf := c.Get("csrf").(string)
	errText := fmt.Sprintf("Error %d: %s (request ID: %s)", code, msg, reqID)

	return html.Page("Error", user, c.Path(), csrf, html.ErrorMain(errText)).Render(c.Response())
}
