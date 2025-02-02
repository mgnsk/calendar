package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// LoginPage render the login page.
func LoginPage(mainTitle string, hasError bool, username, password string) Node {
	return page(mainTitle, "Login", "", nil,
		Form(Class("bg-blue text-center w-full sm:w-1/2 px-3 py-4 mx-auto rounded"), Action("/login"), Method("POST"),
			loginInput("username", "text", "Username", username, hasError),
			loginInput("password", "password", "Password", password, hasError, "my-3"),
			Button(Type("submit"), Class("bg-blue font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Login")),
			If(hasError, P(Class("pt-5 text-red-500 text-sm italic"), Text("Invalid username or password"))),
		),
	)
}

func loginInput(name, typ, placeholder string, value string, hasError bool, extraClasses ...string) Node {
	classes := Classes{
		"border":          true,
		"border-gray-200": true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      hasError,
	}

	for _, class := range extraClasses {
		classes[class] = true
	}

	return Input(classes,
		Name(name),
		Type(typ),
		Placeholder(placeholder),
		Value(value),
		Required(),
	)
}
