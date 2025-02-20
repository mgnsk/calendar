package html

import (
	_ "embed"

	"github.com/mgnsk/calendar"
	"github.com/mgnsk/calendar/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

//go:embed search.js
var searchScript string

//go:embed eventnav.js
var eventNavScript string

//go:embed addevent.js
var addEventScript string

// Page renders a page.
func Page(mainTitle string, user *domain.User, path, csrf string, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    mainTitle,
		Language: "en",
		Head: []Node{
			Link(Rel("icon"), Type("image/x-icon"), Href(calendar.GetAssetPath("favicon.ico"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/easymde/dist/easymde.min.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/@fortawesome/fontawesome-free/css/fontawesome.min.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("app.css"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/htmx.org/dist/htmx.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/mark.js/dist/mark.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/easymde/dist/easymde.min.js"))),
			Script(Raw(eventNavScript)),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			UserNav(user, path, csrf),
			Group(children),
			loadingSpinner(),
			Script(Raw(searchScript)),
			Script(Raw(addEventScript)),
		},
	})
}
