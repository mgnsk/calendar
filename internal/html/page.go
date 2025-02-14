package html

import (
	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// Page renders a page.
func Page(mainTitle string, user *domain.User, path, csrf string, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    mainTitle,
		Language: "en",
		Head: []Node{
			Link(Rel("icon"), Type("image/x-icon"), Href(internal.GetAssetLink("dist/favicon.ico"))),
			Link(Rel("stylesheet"), Href(internal.GetAssetLink("dist/app.css"))),
			ScriptAsyncDefer("dist/htmx.min.js"),
			ScriptRaw(eventNavScript),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			UserNav(user, path, csrf),
			Group(children),
			loadingSpinner(),
			ScriptSync("dist/mark.min.js"),
			ScriptRaw(searchScript),
		},
	})
}
