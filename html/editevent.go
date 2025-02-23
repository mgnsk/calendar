package html

import (
	"net/url"

	"github.com/mgnsk/calendar/contract"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// EditEventMain render the edit event page main content.
func EditEventMain(form contract.EditEventForm, errs url.Values, csrf string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			Form(ID("edit-form"), Class("w-full px-3 py-4 mx-auto"),
				Method("POST"),
				input("title", "text", "Title", form.Title, errs.Get("title"), true),
				input("url", "url", "URL", form.URL, errs.Get("url"), false),
				dateTimeLocalInput("start_at", form.StartAt.String(), errs.Get("start_at"), true),
				input("location", "text", "Location", form.Location, errs.Get("location"), true),
				textarea("desc", form.Description, errs.Get("desc"), true),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Input(Type("hidden"), Name("easymde_cache_key"), Value(form.EventID.String())),
				// TODO: save draft button
				submitButton("Publish"),
			),
		),
	)
}
