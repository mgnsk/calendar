package html

import (
	"net/url"

	"github.com/mgnsk/calendar/contract"
	"github.com/mgnsk/calendar/html/components"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SettingsMain renders the settings page main content.
func SettingsMain(form contract.SettingsForm, errs url.Values) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"),
				Method("POST"),
				Label(Class("block w-full pt-2"), For("title"), Text("Title")),
				components.InputElement("pagetitle", "text", "Title", form.Title, errs.Get("pagetitle"), true, false),

				Label(Class("block w-full pt-2"), For("desc"), Text("Description")),
				components.TextareaElement("pagedesc", form.Description, errs.Get("pagedesc"), 3, false, false),

				Label(Class("block w-full pb-2"), For("stopwords"), Text("Stop words are excluded from tags page. One word per line.")),

				components.TextareaElement("stopwords",
					form.Stopwords,
					"",
					20,
					false,
					false,
				),

				components.SubmitButtonElement("Save"),
			),
		),
	)
}
