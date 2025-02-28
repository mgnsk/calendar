package html

import (
	"net/url"

	"github.com/mgnsk/calendar/contract"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SetupMain renders the setup page main content.
func SetupMain(form contract.SetupForm, errs url.Values, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Method("POST"),
				Label(Class("block w-full pt-2"), For("title"), Text("Title")),
				input("pagetitle", "text", "Title", form.Title, errs.Get("pagetitle"), true, false),

				Label(Class("block w-full pt-2"), For("desc"), Text("Description")),
				textarea("pagedesc", form.Description, errs.Get("pagedesc"), false, false),

				Label(Class("block w-full pt-2"), For("username"), Text("Username")),
				input("username", "text", "Username", form.Username, errs.Get("username"), true, false),

				Label(Class("block w-full pt-2"), For("password1"), Text("Password")),
				input("password1", "password", "Password", form.Password1, errs.Get("password1"), true, false),

				Label(Class("block w-full pt-2"), For("password2"), Text("Password again")),
				input("password2", "password", "Password again", form.Password2, errs.Get("password2"), true, false),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),

				submitButton("Save"),
			),
		),
	)
}
