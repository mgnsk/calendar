package html

import (
	"github.com/mgnsk/calendar/internal"
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func userNav(user *domain.User) Node {
	return Nav(Class("max-w-3xl mx-auto mb-5"),
		Ul(Class("flex justify-between font-semibold flex-row space-x-8"),
			// TODO
			Li(Class("justify-self-start"),
				A(Class("inline-block p-2"), Href("/feed"), rssIcon()),
				A(Class("inline-block p-2"), Href("/ical"), calendarIcon()),
			),

			Iff(user != nil, func() Node {
				return Group{
					Li(Class("justify-self-end"),
						// If(user.Role == domain.Admin,
						// 	A(Class("inline-block p-2"), Href("/users"), Text("Users")),
						// ),
						A(Class("inline-block p-2"), Href("/add"), Text("Add event")),
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

// Page renders a page.
func Page(mainTitle, subTitle string, user *domain.User, children ...Node) Node {
	return HTML5(HTML5Props{
		Title:    mainTitle,
		Language: "en",
		Head: []Node{
			Link(Rel("icon"), Type("image/x-icon"), Href(internal.GetAssetLink("dist/favicon.ico"))),
			Link(Rel("stylesheet"), Href(internal.GetAssetLink("dist/app.css"))),
			ScriptDefer("dist/htmx.min.js"),
			Meta(Name("generator"), Content("Calendar - github.com/mgnsk/calendar")),
		},
		Body: []Node{
			// Attr("onload", "let input = document.getElementById('search'); input ? input.value = '' : false"),
			container(
				userNav(user),
				If(subTitle != "", H1(Class("m-8 text-center text-sm md:text-m"), Text(subTitle))),
				Group(children),
				loadingSpinner(),
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

func input(name, typ, placeholder string, value string, err string, extraClasses ...string) Node {
	classes := Classes{
		"border":          true,
		"border-gray-200": true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      err != "",
	}

	for _, class := range extraClasses {
		classes[class] = true
	}

	return Group{
		If(err != "", P(Class("pt-5 text-red-500 text-sm italic"), Text(err))),
		Input(classes,
			Name(name),
			Type(typ),
			Placeholder(placeholder),
			Value(value),
			Required(),
		),
	}
}

func textarea(name, value string, err string, extraClasses ...string) Node {
	classes := Classes{
		"border":          true,
		"border-gray-200": true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      err != "",
	}

	for _, class := range extraClasses {
		classes[class] = true
	}

	return Group{
		If(err != "", P(Class("pt-5 text-red-500 text-sm italic"), Text(err))),
		Textarea(classes,
			Name(name),
			Text(value),
			Rows("3"),
		),
	}
}

func loadingSpinner() Node {
	return Div(ID("loading-spinner"), Class("my-5 opacity-0 htmx-indicator m-10 mx-auto flex justify-center"),
		spinner(8),
	)
}

func rssIcon() Node {
	return Raw(`<svg class="w-4 h-4" xmlns="http://www.w3.org/2000/svg" width="64" height="64" shape-rendering="geometricPrecision" text-rendering="geometricPrecision" image-rendering="optimizeQuality" fill-rule="evenodd" clip-rule="evenodd" viewBox="0 0 640 640"><path d="M85.206 469.305C38.197 469.305 0 507.632 0 554.345c0 46.95 38.197 84.876 85.206 84.876 47.15 0 85.324-37.926 85.324-84.876 0-46.713-38.162-85.04-85.324-85.04zM.083 217.42v122.683c79.89 0 154.963 31.24 211.514 87.84 56.492 56.434 87.686 131.872 87.686 212.07h123.202c0-232.987-189.57-422.556-422.403-422.556v-.036zM.236-.012v122.706c284.885 0 516.727 232.078 516.727 517.282l123.037.012C640 287.188 352.953 0 .248 0L.236-.012z"/></svg>`)
}

func calendarIcon() Node {
	return Raw(`<svg class="w-5 h-5" version="1.1" id="Layer_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px" y="0px" viewBox="0 0 122.88 122.88" style="enable-background:new 0 0 122.88 122.88" xml:space="preserve"><g><path d="M81.61,4.73c0-2.61,2.58-4.73,5.77-4.73s5.77,2.12,5.77,4.73v20.72c0,2.61-2.58,4.73-5.77,4.73s-5.77-2.12-5.77-4.73V4.73 L81.61,4.73z M77.96,80.76c1.83,0,3.32,1.49,3.32,3.32c0,1.83-1.49,3.32-3.32,3.32l-12.95-0.04l-0.04,12.93 c0,1.83-1.49,3.32-3.32,3.32c-1.83,0-3.32-1.49-3.32-3.32l0.04-12.94L45.44,87.3c-1.83,0-3.32-1.49-3.32-3.32 c0-1.83,1.49-3.32,3.32-3.32l12.94,0.04l0.04-12.93c0-1.83,1.49-3.32,3.32-3.32c1.83,0,3.32,1.49,3.32,3.32l-0.04,12.95 L77.96,80.76L77.96,80.76z M29.61,4.73c0-2.61,2.58-4.73,5.77-4.73s5.77,2.12,5.77,4.73v20.72c0,2.61-2.58,4.73-5.77,4.73 s-5.77-2.12-5.77-4.73V4.73L29.61,4.73z M6.4,45.32h110.08V21.47c0-0.8-0.33-1.53-0.86-2.07c-0.53-0.53-1.26-0.86-2.07-0.86H103 c-1.77,0-3.2-1.43-3.2-3.2c0-1.77,1.43-3.2,3.2-3.2h10.55c2.57,0,4.9,1.05,6.59,2.74c1.69,1.69,2.74,4.02,2.74,6.59v27.06v65.03 c0,2.57-1.05,4.9-2.74,6.59c-1.69,1.69-4.02,2.74-6.59,2.74H9.33c-2.57,0-4.9-1.05-6.59-2.74C1.05,118.45,0,116.12,0,113.55V48.53 V21.47c0-2.57,1.05-4.9,2.74-6.59c1.69-1.69,4.02-2.74,6.59-2.74H20.6c1.77,0,3.2,1.43,3.2,3.2c0,1.77-1.43,3.2-3.2,3.2H9.33 c-0.8,0-1.53,0.33-2.07,0.86c-0.53,0.53-0.86,1.26-0.86,2.07V45.32L6.4,45.32z M116.48,51.73H6.4v61.82c0,0.8,0.33,1.53,0.86,2.07 c0.53,0.53,1.26,0.86,2.07,0.86h104.22c0.8,0,1.53-0.33,2.07-0.86c0.53-0.53,0.86-1.26,0.86-2.07V51.73L116.48,51.73z M50.43,18.54 c-1.77,0-3.2-1.43-3.2-3.2c0-1.77,1.43-3.2,3.2-3.2h21.49c1.77,0,3.2,1.43,3.2,3.2c0,1.77-1.43,3.2-3.2,3.2H50.43L50.43,18.54z"/></g></svg>`)
}
