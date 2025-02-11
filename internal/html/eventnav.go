package html

import (
	"encoding/json"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

type eventNavLink struct {
	Text   string
	URL    string
	Active bool
}

// TODO: if tag filter, tehen somehow show
func eventNav(path string, links []eventNavLink, csrf string) Node {
	return Div(Class("max-w-3xl mx-auto"),
		ScriptRaw(eventNavScript),
		Ul(Class("flex border-b"),
			Map(links, func(link eventNavLink) Node {
				return Li(Classes{
					"flex":           true,
					"items-baseline": true,
					"mr-1":           true,
					"-mb-px":         link.Active,
					"border-l":       link.Active,
					"border-t":       link.Active,
					"border-r":       link.Active,
					"rounded-t":      link.Active,
				},
					A(
						Classes{
							"nav-link":             true,
							"bg-white":             true,
							"inline-block":         true,
							"py-2":                 true,
							"px-2":                 true,
							"md:px-4":              true,
							"text-gray-400":        !link.Active,
							"hover:text-amber-600": !link.Active,
							"text-amber-600":       link.Active,
							"font-semibold":        true,
							"hover:cursor-pointer": true,
						},
						hx.Get(link.URL),
						hx.Trigger("click"),
						hx.On("click", "changeTab(this)"),
						hx.Target("#event-list"),
						hx.Swap("innerHTML"),
						hx.PushURL("true"),
						If(link.Active, Aria("current", "page")),
						Text(link.Text),
						// Href(link.URL),
					),
				)
			}),
			If(path != "/tags",
				Li(Class("flex items-baseline ml-auto border-l border-t border-r rounded-t"),
					Div(Class("relative"),
						Input(Classes{
							// "border":          true,
							// "border-gray-200": true,
							"block":   true,
							"w-full":  true,
							"mx-auto": true,
							"py-2":    true,
							"px-3":    true,
							"rounded": true,
						},
							ID("search"),
							Name("search"),
							Type("text"),
							Placeholder("Filter..."),
							Required(),
							hx.Get(""), // Post to current URL.
							hx.Trigger("keyup delay:0.2s"),
							hx.Target("#event-list"),
							hx.Swap("innerHTML"),
							hx.Indicator("#search-spinner"),
							hx.Vals(string(must(json.Marshal(map[string]string{
								"csrf": csrf,
							})))),
						),
						Div(ID("search-spinner"), Class("opacity-0 absolute top-0 right-0 h-full flex items-center mr-2 htmx-indicator"),
							spinner(2),
						),
					),
				),
			),
		),
	)
}
