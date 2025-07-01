package server

import (
	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/html"
	"maragu.dev/gomponents"
)

// RenderPage renders a HTML page.
func RenderPage(
	c *Context,
	sm *scs.SessionManager,
	content gomponents.Node,
) error {
	// Note: Pop must be before writing headers.
	successMessage := sm.PopString(c.Request().Context(), "flash-success")

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)

	return html.Page(html.PageProps{
		Title:        c.Settings.Title,
		User:         c.User,
		Path:         c.Path(),
		CSRF:         c.CSRF,
		Children:     content,
		FlashSuccess: successMessage,
	}).Render(c.Response())
}
