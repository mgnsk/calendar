//go:build !strictdist

package calendar

import (
	"io/fs"
)

// assetsFS is an empty placeholder assets filesystem for testing.
var assetsFS fs.FS

// GetAssetPath returns the asset path.
func GetAssetPath(name string) string {
	return name
}
