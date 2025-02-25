package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/mgnsk/calendar/html"
	. "maragu.dev/gomponents"
)

// RenderPage renders a HTML page.
func RenderPage(
	c *Context,
	content Node,
) error {
	// Note: Pop must be before writing headers.
	successMessage := c.Session.PopString(c.Request().Context(), "flash-success")

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
