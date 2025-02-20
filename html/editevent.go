package html

import (
	"net/url"

	"github.com/mgnsk/calendar/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// EditEventMain render the edit event page main content.
func EditEventMain(form, errs url.Values, ev *domain.Event, csrf string) Node {
	return Main(
		// TODO: start and end datetime fields and URL field
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("w-full px-3 py-4 mx-auto"),
				Method("POST"),
				input("title", "text", "Title", form, errs, "mb-3"),
				textarea("desc", form, errs),
				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Input(Type("hidden"), Name("event_id"),
					Iff(ev != nil, func() Node {
						return Value(ev.ID.String())
					}),
					If(ev == nil, Value("draft")),
				),
				// TODO: save draft button
				Button(Type("submit"), Class("mt-3 font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Publish")),
			),
		),
	)
}
