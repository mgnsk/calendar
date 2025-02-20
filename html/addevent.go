package html

import (
	"net/url"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// AddEventMain render the add event page main content.
func AddEventMain(form, errs url.Values, csrf string) Node {
	return Main(
		// TODO: start and end datetime fields and URL field
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("w-full px-3 py-4 mx-auto"),
				Action("/add"),
				Method("POST"),
				input("title", "text", "Title", form, errs, "mb-3"),
				textarea("desc", form, errs),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				// TODO: save draft button
				Button(Type("submit"), Class("mt-3 font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Publish")),
			),
		),
	)
}
