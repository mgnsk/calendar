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
	return Nav(Class("max-w-3xl mx-auto"),
		Ul(Class("flex justify-end font-semibold flex-row space-x-8"),
			// Note: when enabling the first link, set parent to justify-between.
			// Li(Class("justify-self-start"),
			// 	A(Class("inline-block p-2"), Href("/"), Text("Events")),
			// ),

			Iff(user != nil, func() Node {
				return Group{
					Li(Class("justify-self-end"),
						// If(user.Role == domain.Admin,
						// 	A(Class("inline-block p-2"), Href("/users"), Text("Users")),
						// ),
						A(Class("inline-block p-2"), Href("/manage"), Text("Manage")),
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

func page(mainTitle, sectionTitle, subTitle string, user *domain.User, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    fmt.Sprintf("%s - %s", mainTitle, sectionTitle),
		Language: "en",
		Head: []Node{
			Link(Rel("icon"), Type("image/x-icon"), Href(internal.GetAssetLink("dist/favicon.ico"))),
			Link(Rel("stylesheet"), Href(internal.GetAssetLink("dist/app.css"))),
			Script(Type("application/javascript"), Defer(), Src(internal.GetAssetLink("dist/htmx.min.js"))),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			Attr("onload", "let input = document.getElementById('search'); input ? input.value = '' : false"),
			container(
				userNav(user),
				H1(Class("m-8 text-center text-xl md:text-4xl font-semibold"), Text(sectionTitle)),
				If(subTitle != "", H1(Class("m-8 text-center text-sm md:text-m"), Text(subTitle))),
				Group(children),
			),
		},
	})
}

func container(children ...Node) Node {
	return Div(Class("m-5 mt-2 px-5"), Group(children))
}

func spinner(size int) Node {
	return Rawf(`<svg class="w-%d h-%d text-gray-300 animate-spin" viewBox="0 0 64 64" fill="none"
     xmlns="http://www.w3.org/2000/svg" width="24" height="24">
     <path
       d="M32 3C35.8083 3 39.5794 3.75011 43.0978 5.20749C46.6163 6.66488 49.8132 8.80101 52.5061 11.4939C55.199 14.1868 57.3351 17.3837 58.7925 20.9022C60.2499 24.4206 61 28.1917 61 32C61 35.8083 60.2499 39.5794 58.7925 43.0978C57.3351 46.6163 55.199 49.8132 52.5061 52.5061C49.8132 55.199 46.6163 57.3351 43.0978 58.7925C39.5794 60.2499 35.8083 61 32 61C28.1917 61 24.4206 60.2499 20.9022 58.7925C17.3837 57.3351 14.1868 55.199 11.4939 52.5061C8.801 49.8132 6.66487 46.6163 5.20749 43.0978C3.7501 39.5794 3 35.8083 3 32C3 28.1917 3.75011 24.4206 5.2075 20.9022C6.66489 17.3837 8.80101 14.1868 11.4939 11.4939C14.1868 8.80099 17.3838 6.66487 20.9022 5.20749C24.4206 3.7501 28.1917 3 32 3L32 3Z"
       stroke="currentColor" stroke-width="5" stroke-linecap="round" stroke-linejoin="round"></path>
     <path
       d="M32 3C36.5778 3 41.0906 4.08374 45.1692 6.16256C49.2477 8.24138 52.7762 11.2562 55.466 14.9605C58.1558 18.6647 59.9304 22.9531 60.6448 27.4748C61.3591 31.9965 60.9928 36.6232 59.5759 40.9762"
       stroke="currentColor" stroke-width="5" stroke-linecap="round" stroke-linejoin="round" class="text-gray-900">
     </path>
</svg>`, size, size)
}
