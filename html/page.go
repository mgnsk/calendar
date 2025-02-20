package html

import (
	_ "embed"
	"strings"
	"text/template"

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
			StyleEl(Raw(faFontStyle)),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/htmx.org/dist/htmx.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/mark.js/dist/mark.min.js"))),
			Script(Defer(), Src(calendar.GetAssetPath("node_modules/easymde/dist/easymde.min.js"))),
			Script(Defer(), Raw(eventNavScript)),
			Script(Defer(), Raw(searchScript)),
			Script(Defer(), Raw(editEventScript)),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			UserNav(user, path, csrf),
			Group(children),
			loadingSpinner(),
		},
	})
}

var faFontStyle string

func init() {
	t := template.Must(template.New("").Parse(`
/*!
 * Font Awesome Free 6.7.2 by @fontawesome - https://fontawesome.com
 * License - https://fontawesome.com/license/free (Icons: CC BY 4.0, Fonts: SIL OFL 1.1, Code: MIT License)
 * Copyright 2024 Fonticons, Inc.
 */
:root, :host {
  --fa-style-family-classic: 'Font Awesome 6 Free';
  --fa-font-solid: normal 900 1em/1 'Font Awesome 6 Free'; }

@font-face {
  font-family: 'Font Awesome 6 Free';
  font-style: normal;
  font-weight: 900;
  font-display: block;
  src: url({{ .woffPath }}) format("woff2"), url({{ .ttfPath }}) format("truetype"); }

.fas,
.fa-solid {
  font-weight: 900; }
`))
	var buf strings.Builder
	if err := t.Execute(&buf, map[string]string{
		"woffPath": calendar.GetAssetPath("node_modules/@fortawesome/fontawesome-free/webfonts/fa-solid-900.woff2"),
		"ttfPath":  calendar.GetAssetPath("node_modules/@fortawesome/fontawesome-free/webfonts/fa-solid-900.ttf"),
	}); err != nil {
		panic(err)
	}

	faFontStyle = buf.String()
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}
