package html

import (
	"fmt"
	"net/http"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ErrorPage renders an error page.
func ErrorPage(mainTitle string, httpCode int, msg, reqID string) Node {
	var s string

	if reqID != "" {
		s = fmt.Sprintf("Error %d: %s (request ID %s)", httpCode, http.StatusText(httpCode), reqID)
	} else {
		s = fmt.Sprintf("Error %d: %s", httpCode, http.StatusText(httpCode))
	}

	return page(mainTitle, s, "", nil,
		Div(Class("px-3 py-4 text-center"),
			P(Text(msg)),
		),
	)
}
