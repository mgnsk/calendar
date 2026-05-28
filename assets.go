package calendar

import (
	"fmt"
	"net/http"
	"time"
)

// RegisterAssetsHandler registers the static assets HTTP handler.
func RegisterAssetsHandler(mux *http.ServeMux) {
	handler := newAssetCacheMiddleware(30 * 24 * time.Hour)(
		http.StripPrefix("/assets", http.FileServerFS(assetsFS)),
	)

	mux.Handle("/assets/", handler)
}

func newAssetCacheMiddleware(d time.Duration) func(http.Handler) http.Handler {
	value := fmt.Sprintf("max-age=%d, immutable", int64(d.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", value)

			next.ServeHTTP(w, r)
		})
	}
}
