package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/domain"
)

func loadSettings(c echo.Context) *domain.Settings {
	if settings, ok := c.Get("settings").(*domain.Settings); ok {
		return settings
	}
	return nil
}

func loadUser(c echo.Context) *domain.User {
	if u, ok := c.Get("user").(*domain.User); ok {
		return u
	}
	return nil
}
