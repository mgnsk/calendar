package html

import (
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// StopWordsMain renders the stop words form.
func StopWordsMain(words []string, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full  px-3 py-4 mx-auto"),
				Method("POST"),

				Label(Class("block w-full pb-2"), For("words"), Text("Stop words are excluded from tags page. One word per line.")),
				Textarea(baseInputClasses(false),
					Name("words"),
					ID("words"),
					Text(strings.Join(words, "\n")),
					Rows("20"),
				),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),

				submitButton("Save"),
			),
		),
	)
}
