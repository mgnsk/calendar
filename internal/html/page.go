package html

import (
	_ "embed"

	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

//go:embed search.js
var searchScript string

//go:embed eventnav.js
var eventNavScript string

// Page renders a page.
func Page(mainTitle string, user *domain.User, path, csrf string, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    mainTitle,
		Language: "en",
		Head: []Node{
			Link(Rel("icon"), Type("image/x-icon"), Href(internal.GetAssetLink("dist/favicon.ico"))),
			Link(Rel("stylesheet"), Href(internal.GetAssetLink("dist/app.css"))),
			Script(Async(), Defer(), Src(internal.GetAssetLink("dist/htmx.min.js"))),
			Script(Defer(), Src("dist/mark.min.js")),
			Script(Raw(eventNavScript)),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			UserNav(user, path, csrf),
			Group(children),
			loadingSpinner(),
			Script(Raw(searchScript)),
		},
	})
}
