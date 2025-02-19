//go:build !strictdist

package calendar

import (
	"io/fs"
)

// AssetsFS is an empty placeholder assets filesystem for testing.
var AssetsFS fs.FS
