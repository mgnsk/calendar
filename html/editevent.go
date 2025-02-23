package html

import (
	"net/url"
	"strconv"

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
				dateTimeLocalInput("start_at", form.StartAt, errs.Get("start_at"), true),

				Div(Class("relative"),
					input("location", "text", "Location", form.Location, errs.Get("location"), true),
					Div(ID("location-spinner"), Class("opacity-0 absolute top-0 right-0 h-full flex items-center mr-2"),
						spinner(2),
					),
				),

				textarea("desc", form.Description, errs.Get("desc"), true),

				Input(Type("hidden"), Name("csrf"), Value(csrf)),
				Input(Type("hidden"), Name("easymde_cache_key"), Value(form.EventID.String())),
				Input(Type("hidden"), Name("latitude"), Value(strconv.FormatFloat(form.Latitude, 'f', -1, 64))),
				Input(Type("hidden"), Name("longitude"), Value(strconv.FormatFloat(form.Longitude, 'f', -1, 64))),
				Input(Type("hidden"), Name("user_timezone")),
				Script(Raw(`document.querySelector('[name="user_timezone"]').value = Intl.DateTimeFormat().resolvedOptions().timeZone;`)),
				// TODO: save draft button
				submitButton("Publish"),
			),
		),
	)
}
