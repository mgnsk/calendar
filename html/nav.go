package html

import (
	"encoding/json"

	"github.com/mgnsk/calendar/domain"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// UserNav renders the user navigation.
func UserNav(user *domain.User, children Node) Node {
	return Nav(Class("sticky top-0 bg-white max-w-3xl mx-auto z-1"),
		Ul(Class("flex justify-between font-semibold flex-row space-x-8 mb-5"),
			// TODO: find better icons
			Li(Class("justify-self-start align-start"),
				A(Class("inline-block p-2"), Href("/"), Text("Home")),
				A(Class("inline-block p-2"), Title("RSS feed"), Href("/feed"), rssIcon()),
				A(Class("inline-block p-2"), Title("iCal URL"), ID("ical-link"), calendarIcon()),
				A(Class("inline-block p-2"), Title("Add to Google Calendar"), ID("google-calendar-link"), Target("_blank"), calendarIcon()),
				Script(Raw(`document.getElementById("ical-link").setAttribute("href", "webcals://" + window.location.host + "/calendar.ics")`)),
				Script(Raw(`document.getElementById("google-calendar-link").setAttribute("href", "https://calendar.google.com/calendar/render?cid=" + window.location.protocol + "//" + window.location.host + "/calendar.ics")`)),
			),

			Iff(user != nil, func() Node {
				return Group{
					Li(Class("justify-self-end"),
						A(Class("inline-block p-2"), Href("/edit/0"), Text("Add event")),
						If(user.Role == domain.Admin,
							A(Class("inline-block p-2"), Href("/users"), Text("Users")),
						),
						A(Class("inline-block p-2"), Href("/logout"), Text("Logout")),
					),
				}
			}),

			If(user == nil,
				Li(Class("justify-self-end"),
					A(Class("inline-block p-2"), Href("/login"), Text("Login")),
				),
			),
		),
		children,
	)
}

type eventNavLink struct {
	Text   string
	URL    string
	Active bool
}

// EventNav renders the event navigation.
func EventNav(user *domain.User, path, csrf string) Node {
	links := []eventNavLink{
		{
			Text:   "Latest",
			URL:    "/",
			Active: path == "/",
		},
		{
			Text:   "Upcoming",
			URL:    "/upcoming",
			Active: path == "/upcoming",
		},
		{
			Text:   "Past",
			URL:    "/past",
			Active: path == "/past",
		},
		{
			Text:   "Tags",
			URL:    "/tags",
			Active: path == "/tags",
		},
	}

	if user != nil {
		links = append(links, eventNavLink{
			Text:   "My events",
			URL:    "/my-events",
			Active: path == "/my-events",
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
						hx.Post(link.URL),
						hx.Trigger("click"),
						If(link.URL == "/tags", hx.On("click", "changeTab(this); setSearch('')")), // Clear search when clicking tags tab.
						If(link.URL != "/tags", hx.On("click", "changeTab(this)")),                // Keep search query when clicking event tabs.
						If(link.URL != "/tags", hx.Include("[name='search']")),                    // Keep search query when clicking event tabs.
						hx.Target("#event-list"),
						hx.Swap("innerHTML"),
						hx.PushURL("true"),
						hx.Indicator("#loading-spinner"),
						hx.Vals(string(must(json.Marshal(map[string]string{
							"csrf": csrf,
						})))),
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
						hx.Post(""), // Post to current URL.
						hx.Trigger("input delay:0.2s"),
						hx.Target("#event-list"),
						hx.Swap("innerHTML"),
						hx.Indicator("#search-spinner, #loading-spinner"),
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
	)
}
