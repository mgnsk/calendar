package html

import (
	"net/url"

	"github.com/mgnsk/calendar/pkg/snowflake"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// EditEventMain render the edit event page main content.
func EditEventMain(form, errs url.Values, eventID snowflake.ID, csrf string) Node {
	return Main(
		// TODO: start and end datetime fields and URL field
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("w-full px-3 py-4 mx-auto"),
				Method("POST"),
				input("title", "text", "Title", form, errs, "mb-3"),
				input("url", "url", "URL", form, errs, "mb-3"),
				Div(Class("flex items-center"),
					dateTimeLocalInput("start_at", form, errs, "mb-3"),
					Span(Class("px-5"), Text("until")),
					dateTimeLocalInput("end_at", form, errs, "mb-3"),
				),
				textarea("desc", form, errs),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Input(Type("hidden"), Name("event_id"), Value(eventID.String())),
				// TODO: save draft button
				Button(Type("submit"), Class("mt-3 font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Publish")),
			),
		),
	)
}
