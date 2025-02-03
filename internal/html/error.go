package html

import (
	"fmt"
	"net/http"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ErrorPage renders an error page.
func ErrorPage(mainTitle string, httpCode int, msg string) Node {
	return page(mainTitle, fmt.Sprintf("Error %d: %s", httpCode, http.StatusText(httpCode)), "", nil,
		Div(Class("px-3 py-4 text-center"),
			P(Text(msg)),
		),
	)
}
