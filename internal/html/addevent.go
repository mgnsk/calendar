package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// AddEventPage render the add event page.
func AddEventPage(mainTitle string, errs map[string]string, title, csrf string) Node {
	return Page(mainTitle, "Add event", nil,
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Action("/add"),
				Method("POST"),
				input("title", "text", "Title", title, errs["title"]),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Button(Type("submit"), Class("font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Submit")),
			),
		),
	)
}
