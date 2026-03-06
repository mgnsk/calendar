package components

import (
	"maps"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

// BaseFormElementClasses is base form element classes.
// TODO: refactor
func BaseFormElementClasses() Classes {
	return Classes{
		"py-2":    true,
		"px-3":    true,
		"rounded": true,
		"mb-3":    true,
		"block":   true,
		"w-full":  true,
		"mx-auto": true,
	}
}

func baseInputClasses(hasError bool) Classes {
	classes := BaseFormElementClasses()
	maps.Copy(classes, Classes{
		"border":          true,
		"border-gray-200": true,
		"bg-red-100":      hasError,
	})

	return classes
}

func withErrors(err string, input Node) Node {
	return Group{
		If(err != "", P(Class("text-red-500 text-sm italic"), Text(err))),
		input,
	}
}

func buttonClasses() Classes {
	classes := baseInputClasses(false)
	maps.Copy(classes, Classes{
		"mt-3":                 true,
		"font-bold":            true,
		"hover:cursor-pointer": true,
		"hover:bg-amber-600/5": true,
	})

	return classes
}

func must[V any](v V, err error) V {
	if err != nil {
		panic(err)
	}
	return v
}
