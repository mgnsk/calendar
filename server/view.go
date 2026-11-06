package server

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/webkit/session"
	"github.com/mgnsk/webkit/web"
	"maragu.dev/gomponents"
)

// RenderPage renders a HTML page.
func RenderPage(
	w http.ResponseWriter,
	r *http.Request,
	sm *scs.SessionManager,
	content gomponents.Node,
) {
	// Note: Pop must be before writing headers.
	successMessage := session.PopFlash(r.Context(), sm)

	title := "TODO: nil settings"
	if settings := GetSettings(r.Context()); settings != nil {
		title = settings.Title
	}

	web.Render(w, html.Page(html.PageProps{
		Title:        title,
		User:         UserKey.Get(r.Context()),
		Path:         r.URL.Path,
		Children:     content,
		FlashSuccess: successMessage,
	}))
}
