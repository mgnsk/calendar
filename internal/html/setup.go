package html

import (
	"net/url"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SetupMain renders the setup page main content.
func SetupMain(form, errs url.Values, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Action("/setup"),
				Method("POST"),
				Label(Class("block w-full pt-2"), For("title"), Text("Title")),
				input("title", "text", "Title", form, errs, "mb-3"),

				Label(Class("block w-full pt-2"), For("desc"), Text("Description")),
				textarea("desc", form, errs, "mb-3"),

				Label(Class("block w-full pt-2"), For("username"), Text("Username")),
				input("username", "text", "Username", form, errs, "mb-3"),

				Label(Class("block w-full pt-2"), For("password1"), Text("Password")),
				input("password1", "password", "Password", form, errs, "mb-3"),

				Label(Class("block w-full pt-2"), For("password2"), Text("Password again")),
				input("password2", "password", "Password again", form, errs, "mb-3"),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Button(Type("submit"), Class("font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Save")),
			),
		),
	)
}
