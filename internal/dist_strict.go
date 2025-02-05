//go:build strictdist

package internal

import (
	"embed"
)

// DistFS contains the bundled assets for web page.
//
//go:embed dist/app.css
//go:embed dist/htmx.min.js
//go:embed dist/mark.min.js
//go:embed dist/favicon.ico
var DistFS embed.FS
