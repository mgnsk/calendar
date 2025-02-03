package html

import (
	"github.com/mgnsk/calendar/internal/domain"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// SetupPage render the setup page.
// TODO: validation of separate fields.
func SetupPage(hasError bool, s *domain.Settings) Node {
	return page("Events setup", "Setup", "", nil,
		Div(Class("max-w-3xl mx-auto"),
			Form(Class("text-center w-full sm:w-1/2 px-3 py-4 mx-auto"), Action("/login"), Method("POST"),
				input("page_title", "text", "Page Title", s.Title, hasError),
				input("description", "description", "Description", s.Description, hasError, "my-3"),
				input("base_url", "base_url", "Base URL", s.BaseURL.String(), hasError, "my-3"),
				Button(Type("submit"), Class("font-bold py-2 px-4 rounded border block mx-auto w-full hover:bg-amber-600 hover:bg-opacity-5"), Text("Login")),
				If(hasError, P(Class("pt-5 text-red-500 text-sm italic"), Text("All fields are required"))),
			),
		),
	)
}
