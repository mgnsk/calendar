package html

import (
	"fmt"

	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func userNav(user *domain.User) Node {
	return Nav(
		Ul(Class("flex justify-between font-semibold flex-row space-x-8"),
			Li(Class("justify-self-start"),
				A(Class("inline-block p-2"), Href("/"), Text("Events")),
			),

			Iff(user != nil, func() Node {
				return Group{
					Li(Class("justify-self-end"),
						If(user.Role == domain.Admin,
							A(Class("inline-block p-2"), Href("/users"), Text("Users")),
						),
						A(Class("inline-block p-2"), Href("/change-password"), Text("Change password")),
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
	)
}

func page(mainTitle, sectionTitle string, user *domain.User, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    fmt.Sprintf("%s - %s", mainTitle, sectionTitle),
		Language: "en",
		Head: []Node{
			Link(Rel("stylesheet"), Href(fmt.Sprintf("/dist/app.css?crc=%d", internal.Checksums["dist/app.css"]))),
		},
		Body: []Node{
			container(
				userNav(user),
				H1(Class("m-8 text-center text-4xl font-semibold"), Text(sectionTitle)),
				Group(children),
			),
		},
	})
}

func container(children ...Node) Node {
	return Div(Class("m-5 mt-2 max-w-3xl mx-auto px-5"), Group(children))
}
