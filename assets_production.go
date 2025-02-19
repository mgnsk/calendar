//go:build strictdist

package calendar

import (
	"embed"
)

// DistFS contains the bundled assets for web page.
//
//go:embed favicon.ico
//go:embed app.css
//go:embed node_modules/htmx.org/dist/htmx.min.js
//go:embed node_modules/mark.js/dist/mark.min.js
var DistFS embed.FS
