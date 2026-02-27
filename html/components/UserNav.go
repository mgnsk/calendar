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
			// TODO: find better icons
			Li(Class("justify-self-start align-start"),
				A(Class("inline-block p-2"), Href("/"), Text("Home")),
				A(Class("inline-block p-2"), Title("RSS feed"), Href("/feed"), rssIcon()),
				A(Class("inline-block p-2"), Title("iCal URL"), ID("ical-link"), calendarIcon()),
				A(Class("inline-block p-2"), Title("Add to Google Calendar"), ID("google-calendar-link"), Target("_blank"), calendarIcon()),
				Script(Raw(`document.getElementById("ical-link").setAttribute("href", "webcal://" + window.location.host + "/calendar.ics")`)),
				Script(Raw(`document.getElementById("google-calendar-link").setAttribute("href", "https://calendar.google.com/calendar/render?cid=" + "http://" + window.location.host + "/calendar.ics")`)),
			),

			Iff(user != nil, func() Node {
				return Group{
					Li(Class("justify-self-end"),
						A(Class("inline-block p-2"), Href("/edit/0"), Text("Add event")),
						If(user.Role == domain.Admin, Group{
							A(Class("inline-block p-2"), Href("/stopwords"), Text("Stop words"), Title("Configure tag cloud stop words")),
							A(Class("inline-block p-2"), Href("/users"), Text("Users"), Title("Manage users")),
						}),
						A(Class("inline-block p-2"), Href("/logout"), Text("Logout")),
					),
				}
			}),

			Iff(user == nil, func() Node {
				return Li(Class("justify-self-end"),
					A(Class("inline-block p-2"), Href("/login"), Text("Login")),
				)
			}),
		),
		children,
	)
}
