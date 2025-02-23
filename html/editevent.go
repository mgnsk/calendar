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
		Div(Class("max-w-3xl mx-auto"),
			Form(ID("edit-form"), Class("w-full px-3 py-4 mx-auto"),
				Method("POST"),
				input("title", "text", "Title", form, errs),
				input("url", "url", "URL", form, errs),
				Div(Class("flex items-center"),
					dateTimeLocalInput("start_at", form, errs),
					Span(Class("px-5"), Text("until")),
					dateTimeLocalInput("end_at", form, errs),
				),
				textarea("desc", form, errs),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Input(Type("hidden"), Name("event_id"), Value(eventID.String())),
				// TODO: save draft button
				submitButton("Publish"),
			),
		),
	)
}
