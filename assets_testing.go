//go:build !strictdist

package calendar

import (
	"io/fs"
)

// assetsFS is an empty placeholder assets filesystem for testing.
// Asset paths are returned unchanged and nothing is served.
var assetsFS fs.FS
