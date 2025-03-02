package html

import (
	_ "embed"
	"fmt"

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

//go:embed editevent.js
var editEventScript string

// PageProps is props for page.
type PageProps struct {
	Title        string
	User         *domain.User
	Path         string
	CSRF         string
	Children     Node
	FlashSuccess string
}

// Page renders a page.
func Page(props PageProps) Node {
	return HTML5(HTML5Props{
		Title:    props.Title,
		Language: "en",
		Head: []Node{
			Link(Rel("alternate"), Type("application/rss+xml"), Title(fmt.Sprintf("RSS feed for %s", props.Title)), Href("/feed")),
			Link(Rel("icon"), Type("image/x-icon"), Href(calendar.GetAssetPath("favicon.ico"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/easymde/dist/easymde.min.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/@fortawesome/fontawesome-free/css/fontawesome.min.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/@fortawesome/fontawesome-free/css/solid.min.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/leaflet/dist/leaflet.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/leaflet-geosearch/dist/geosearch.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("node_modules/jquery-ui/dist/themes/base/jquery-ui.min.css"))),
			Link(Rel("stylesheet"), Href(calendar.GetAssetPath("app.css"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/htmx.org/dist/htmx.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/mark.js/dist/mark.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/easymde/dist/easymde.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/leaflet/dist/leaflet.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/leaflet-geosearch/dist/bundle.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/jquery/dist/jquery.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/jquery-ui/dist/jquery-ui.min.js"))),
			Script(Defer(), Raw(eventNavScript)),
			Script(Defer(), Raw(searchScript)),
			Script(Defer(), Raw(editEventScript)),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			UserNav(
				props.User,
				If(
					props.Path == "/" ||
						props.Path == "/upcoming" ||
						props.Path == "/past" ||
						props.Path == "/tags" ||
						props.Path == "/my-events",
					EventNav(props.User, props.Path, props.CSRF),
				),
			),
			props.Children,
			loadingSpinner(),
			If(props.FlashSuccess != "", flashMessage(true, props.FlashSuccess)),
		},
	})
}

func flashMessage(success bool, message string) Node {
	return Div(
		Div(ID("alert"), Classes{
			"fixed":           true,
			"bottom-5":        true,
			"right-5":         true,
			"bg-teal-100":     success,
			"bg-red-100":      !success,
			"border-t-4":      true,
			"border-teal-500": success,
			"border-red-500":  !success,
			"rounded-b":       true,
			"text-teal-900":   success,
			"text-red-900":    !success,
			"px-4":            true,
			"py-3":            true,
			"shadow-md":       true,
		},
			Role("alert"),
			Div(Class("flex items-center"),
				I(Class("fa fa-info-circle pr-1"), Aria("hidden", "true")),
				P(Class("font-bold"), Text(message)),
			),
		),
		Script(Raw(`setTimeout(() => document.getElementById("alert").remove(), 5000)`)),
	)
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}
