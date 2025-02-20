//go:build strictdist

package calendar

import (
	"embed"
	"fmt"
	"hash/crc32"
	"io/fs"
)

// assetsFS contains the bundled assets for web page.
//
//go:embed favicon.ico
//go:embed app.css
//go:embed node_modules/htmx.org/dist/htmx.min.js
//go:embed node_modules/mark.js/dist/mark.min.js
//go:embed node_modules/easymde/dist/easymde.min.js
//go:embed node_modules/easymde/dist/easymde.min.css
//go:embed node_modules/@fortawesome/fontawesome-free/css/fontawesome.min.css
var assetsFS embed.FS

// GetAssetPath returns the asset path with appended
// CRC checksum query parameter for cache-busting.
func GetAssetPath(name string) string {
	b, err := fs.ReadFile(assetsFS, name)
	if err != nil {
		panic(err)
	}

	sum := crc32.ChecksumIEEE(b)

	return fmt.Sprintf("assets/%s?crc=%d", name, sum)
}
