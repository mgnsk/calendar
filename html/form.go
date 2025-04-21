package html

import (
	"fmt"
	"maps"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

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

func submitButton(text string) Node {
	return Button(buttonClasses(),
		Type("submit"),
		Text(text),
	)
}

func input(name, typ, placeholder string, value, err string, required, autocomplete bool) Node {
	return withErrors(err,
		Input(baseInputClasses(err != ""),
			Name(name),
			Type(typ),
			If(required, Placeholder(fmt.Sprintf("%s*", placeholder))),
			If(!required, Placeholder(placeholder)),
			Value(value),
			If(required, Required()),
			If(!autocomplete, AutoComplete("off")),
		),
	)
}

func textarea(name string, value, err string, required, autocomplete bool) Node {
	return withErrors(err,
		Textarea(baseInputClasses(err != ""),
			Name(name),
			Text(value),
			Rows("3"),
			If(required, Required()),
			If(!autocomplete, AutoComplete("off")),
		),
	)
}

func dateTimeLocalInput(name string, value, err string, required, autocomplete bool) Node {
	return withErrors(err,
		Input(baseInputClasses(err != ""),
			Name(name),
			Type("datetime-local"),
			Value(value),
			If(required, Required()),
			If(!autocomplete, AutoComplete("off")),
		),
	)
}

func baseFormElementClasses() Classes {
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
	classes := baseFormElementClasses()
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
