package html

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ErrorMain renders the error page main content.
func ErrorMain(text string) Node {
	return Main(
		Div(Class("max-w-3xl mx-auto"),
			H1(Text(text)),
		),
	)
}
