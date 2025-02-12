package html

import (
	_ "embed"

	"github.com/mgnsk/calendar/internal"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

//go:embed search.js
var searchScript string

//go:embed eventnav.js
var eventNavScript string

// ScriptSync creates a synchronously loaded script.
func ScriptSync(filename string) Node {
	return Script(Src(internal.GetAssetLink(filename)))
}

// ScriptDefer creates a deferred script.
func ScriptDefer(filename string) Node {
	return Script(Defer(), Src(internal.GetAssetLink(filename)))
}

// ScriptRaw creates a raw script.
func ScriptRaw(script string) Node {
	return Script(Raw(script))
}
