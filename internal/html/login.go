package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// LoginPage render the login page.
func LoginPage(mainTitle string, errs map[string]string, username, password string) Node {
	return page(mainTitle, "Login", "", nil,
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"), Action("/login"), Method("POST"),
				input("username", "text", "Username", username, errs["username"]),
				input("password", "password", "Password", password, errs["password"], "my-3"),
				Button(Type("submit"), Class("font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Login")),
			),
		),
	)
}
