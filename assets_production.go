//go:build strictdist

package calendar

import (
	"embed"
)

// AssetsFS contains the bundled assets for web page.
//
//go:embed favicon.ico
//go:embed app.css
//go:embed node_modules/htmx.org/dist/htmx.min.js
//go:embed node_modules/mark.js/dist/mark.min.js
var AssetsFS embed.FS
