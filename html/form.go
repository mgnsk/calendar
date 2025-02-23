package html

import (
	"maps"
	"net/url"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func submitButton(text string) Node {
	classes := baseInputClasses("", nil)
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

func input(name, typ, placeholder string, form, errs url.Values) Node {
	return withErrors(name, errs,
		Input(baseInputClasses(name, errs),
			Name(name),
			Type(typ),
			Placeholder(placeholder),
			Value(form.Get(name)),
			Required(),
		),
	)
}

func textarea(name string, form, errs url.Values) Node {
	return withErrors(name, errs,
		Textarea(baseInputClasses(name, errs),
			Name(name),
			Text(form.Get(name)),
			Rows("3"),
			Required(),
		),
	)
}

func dateTimeLocalInput(name string, form, errs url.Values) Node {
	return withErrors(name, errs,
		Input(baseInputClasses(name, errs),
			Name(name),
			Type("datetime-local"),
			Value(form.Get(name)),
			Required(),
		),
	)
}

func baseInputClasses(name string, errs url.Values) Classes {
	return Classes{
		"border":          true,
		"border-gray-200": true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      errs.Has(name),
		"mb-3":            true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
	}
}

func withErrors(name string, errs url.Values, input Node) Node {
	return Group{
		If(errs.Has(name), P(Class("text-red-500 text-sm italic"), Text(errs.Get(name)))),
		input,
	}
}

// DateTimeFormat is the datetime-local input format.
const DateTimeFormat = "2006-01-02T15:04"
