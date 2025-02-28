package html

import (
	"maps"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func submitButton(text string) Node {
	classes := baseInputClasses(false)
	maps.Copy(classes, Classes{
		"mt-3":                 true,
		"font-bold":            true,
		"hover:cursor-pointer": true,
		"hover:bg-amber-600/5": true,
	})

	return Button(classes,
		Type("submit"),
		Text(text),
	)
}

func input(name, typ, placeholder string, value, err string, required, autocomplete bool) Node {
	return withErrors(err,
		Input(baseInputClasses(err != ""),
			Name(name),
			Type(typ),
			Placeholder(placeholder),
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

func baseInputClasses(hasError bool) Classes {
	return Classes{
		"border":          true,
		"border-gray-200": true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      hasError,
		"mb-3":            true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
	}
}

func withErrors(err string, input Node) Node {
	return Group{
		If(err != "", P(Class("text-red-500 text-sm italic"), Text(err))),
		input,
	}
}
