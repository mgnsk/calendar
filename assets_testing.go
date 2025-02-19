//go:build !strictdist

package calendar

import (
	"io/fs"
)

// DistFS is an empty placeholder assets filesystem for testing.
var DistFS fs.FS
