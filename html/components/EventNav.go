package components

import (
	"github.com/mgnsk/calendar/domain"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// EventNav renders the event navigation.
func EventNav(user *domain.User, currentPath string) Node {
	type eventNavLink struct {
		Text   string
		URL    string
		Active bool
	}

	links := []eventNavLink{
		{
			Text:   "Upcoming",
			URL:    "/",
			Active: currentPath == "/",
		},
		{
			Text:   "Past",
			URL:    "/past",
			Active: currentPath == "/past",
		},
	}

	if user != nil {
		links = append(links, eventNavLink{
			Text:   "My events",
			URL:    "/my-events",
			Active: currentPath == "/my-events",
		})
	}

	return Div(Class("max-w-3xl mx-auto"),
		Ul(Class("flex border-b border-gray-200"),
			Map(links, func(link eventNavLink) Node {
				return Li(Classes{
					"flex":            true,
					"items-baseline":  true,
					"mr-1":            true,
					"border-gray-200": true,
					"-mb-px":          link.Active,
					"border-l":        link.Active,
					"border-t":        link.Active,
					"border-r":        link.Active,
					"rounded-t":       link.Active,
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
						Href(link.URL),
						If(link.Active, Aria("current", "page")),
						Text(link.Text),
					),
				)
			}),
			Li(Class("flex items-baseline ml-auto border-l border-t border-r border-gray-200 rounded-t"),
				Div(Class("relative"),
					Input(Classes{
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
						hx.Get(""), // Post to current URL.
						hx.Trigger("input delay:0.2s"),
						hx.Target("#event-list"),
						hx.Swap("innerHTML"),
						hx.Indicator("#search-spinner, #loading-spinner"),
					),
					Div(ID("search-spinner"), Class("opacity-0 absolute top-0 right-0 h-full flex items-center mr-2 htmx-indicator"),
						Spinner(2),
					),
				),
			),
		),
	)
}
