package html

import (
	"net/url"

	"github.com/mgnsk/calendar/contract"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// LoginMain renders the login page main content.
func LoginMain(form contract.LoginForm, errs url.Values, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Method("POST"),
				input("username", "text", "Username", form.Username, errs.Get("username")),
				input("password", "password", "Password", form.Password, errs.Get("password")),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				submitButton("Login"),
			),
		),
	)
}
