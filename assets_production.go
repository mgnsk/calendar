//go:build strictdist

package calendar

import (
	"embed"
)

// assetsFS contains the bundled assets for web page.
//
//go:embed favicon.ico
//go:embed app.css
//go:embed node_modules/htmx.org/dist/htmx.min.js
//go:embed node_modules/mark.js/dist/mark.min.js
//go:embed node_modules/easymde/dist
//go:embed node_modules/@fortawesome/fontawesome-free/css/fontawesome.min.css
//go:embed node_modules/@fortawesome/fontawesome-free/css/solid.min.css
//go:embed node_modules/@fortawesome/fontawesome-free/css/regular.min.css
//go:embed node_modules/@fortawesome/fontawesome-free/webfonts/fa-solid-900.woff2
//go:embed node_modules/@fortawesome/fontawesome-free/webfonts/fa-regular-400.woff2
//go:embed node_modules/leaflet/dist
//go:embed node_modules/leaflet-geosearch/dist
//go:embed node_modules/jquery/dist/jquery.min.js
//go:embed node_modules/jquery-ui/dist/jquery-ui.min.js
//go:embed node_modules/jquery-ui/dist/themes/base
var assetsFS embed.FS
