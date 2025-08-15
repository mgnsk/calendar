package html

import (
	"strings"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// StopWordsMain renders the stop words form.
func StopWordsMain(words []string, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl px-3 py-4 mx-auto"),
			P(Class("pb-2"), Text("Stop words are excluded from tags page. One word per line.")),

			Form(Class("text-center w-full mx-auto"),
				Method("POST"),

				Textarea(baseInputClasses(false),
					Name("words"),
					Text(strings.Join(words, "\n")),
					Rows("20"),
				),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),

				submitButton("Save"),
			),
		),
	)
}
