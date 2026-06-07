package components

import (
	"github.com/mgnsk/calendar/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// UserNav renders the user navigation.
func UserNav(user *domain.User, children Node) Node {
	return Nav(Class("sticky top-0 bg-white max-w-3xl mx-auto z-1"),
		Ul(Class("flex justify-between font-semibold flex-row space-x-8 mb-5"),
			Li(Class("justify-self-start align-start"),
				A(Class("inline-block p-2"), Href("/"), Title("Home"),
					I(Class("fa-solid fa-house pr-1"), Aria("hidden", "true")),
				),
				A(Class("inline-block p-2"), Title("RSS feed"), Href("/feed"),
					I(Class("fa-solid fa-rss pr-1"), Aria("hidden", "true")),
				),
				A(Class("inline-block p-2"), Title("iCal URL"), ID("ical-link"),
					I(Class("fa-regular fa-calendar pr-1"), Aria("hidden", "true")),
				),
				A(Class("inline-block p-2"), Title("Add to Google Calendar"), ID("google-calendar-link"), Target("_blank"),
					I(Class("fa-regular fa-calendar-plus pr-1"), Aria("hidden", "true")),
				),
				Script(Raw(`window.webcalURL = "webcal://" + window.location.host + "/calendar.ics"`)),
				Script(Raw(`document.getElementById("ical-link").setAttribute("href", window.webcalURL)`)),
				Script(Raw(`document.getElementById("google-calendar-link").setAttribute("href", "https://calendar.google.com/calendar/render?cid=" + window.webcalURL)`)),
			),

			Iff(user != nil, func() Node {
				return Group{
					Li(Class("justify-self-end"),
						A(Class("inline-block p-2"), Href("/edit/0"), Title("Add event"),
							I(Class("fa-solid fa-plus pr-1"), Aria("hidden", "true")),
						),
						If(user.Role == domain.Admin, Group{
							A(Class("inline-block p-2"), Href("/settings"), Title("Settings"),
								I(Class("fa-solid fa-gear pr-1"), Aria("hidden", "true")),
							),
							A(Class("inline-block p-2"), Href("/users"), Title("Manage users"),
								I(Class("fa-solid fa-users pr-1"), Aria("hidden", "true")),
							),
						}),
						A(Class("inline-block p-2"), Href("/logout"), Title("Logout"),
							I(Class("fa-solid fa-arrow-right-from-bracket pr-1"), Aria("hidden", "true")),
						),
					),
				}
			}),

			Iff(user == nil, func() Node {
				return Li(Class("justify-self-end"),
					A(Class("inline-block p-2"), Href("/login"), Title("Login"),
						I(Class("fa-solid fa-arrow-right-to-bracket pr-1"), Aria("hidden", "true")),
					),
				)
			}),
		),
		children,
	)
}
