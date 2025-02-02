//go:build test

package internal

import (
	"embed"
)

// DistFS contains the bundled assets for web page.
//
//go:embed dist/*
var DistFS embed.FS
