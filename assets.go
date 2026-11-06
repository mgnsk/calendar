package calendar

import (
	"net/http"

	"github.com/mgnsk/webkit/assets"
)

// Assets are the static assets.
var Assets = assets.New(assetsFS)

// GetAssetPath returns the asset path with appended
// CRC checksum query parameter for cache-busting.
func GetAssetPath(name string) string {
	return Assets.Path(name)
}

// RegisterAssetsHandler registers the static assets HTTP handler.
func RegisterAssetsHandler(mux *http.ServeMux) {
	Assets.Register(mux)
}
