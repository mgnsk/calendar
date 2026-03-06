package html

import (
	"strings"

	"github.com/mgnsk/calendar/html/components"
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

				components.TextareaElement("words",
					strings.Join(words, "\n"),
					"",
					20,
					false,
					false,
				),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),

				components.SubmitButtonElement("Save"),
			),
		),
	)
}
