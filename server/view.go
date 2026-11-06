package server

import (
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/mgnsk/calendar/html"
	"github.com/mgnsk/webkit/session"
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

	settings := GetSettings(r.Context())
	user := GetUser(r.Context())

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")

	title := "TODO: nil settings"
	if settings != nil {
		title = settings.Title
	}

	if err := html.Page(html.PageProps{
		Title:        title,
		User:         user,
		Path:         r.URL.Path,
		Children:     content,
		FlashSuccess: successMessage,
	}).Render(w); err != nil {
		panic(err)
	}
}
