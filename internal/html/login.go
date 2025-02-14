package html

import (
	"net/url"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// LoginMain renders the login page main content.
func LoginMain(form, errs url.Values, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Action("/login"),
				Method("POST"),
				input("username", "text", "Username", form, errs),
				input("password", "password", "Password", form, errs, "my-3"),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Button(Type("submit"), Class("font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Login")),
			),
		),
	)
}
