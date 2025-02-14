package html

import (
	"net/url"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/components"
	. "maragu.dev/gomponents/html"
)

func input(name, typ, placeholder string, form, errs url.Values, extraClasses ...string) Node {
	classes := Classes{
		"border":          true,
		"border-gray-200": true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      errs.Has(name),
	}

	for _, class := range extraClasses {
		classes[class] = true
	}

	return Group{
		If(errs.Has(name), P(Class("pt-5 text-red-500 text-sm italic"), Text(errs.Get(name)))),
		Input(classes,
			Name(name),
			Type(typ),
			Placeholder(placeholder),
			Value(form.Get(name)),
			Required(),
		),
	}
}

func textarea(name string, form, errs url.Values, extraClasses ...string) Node {
	classes := Classes{
		"border":          true,
		"border-gray-200": true,
		"block":           true,
		"w-full":          true,
		"mx-auto":         true,
		"py-2":            true,
		"px-3":            true,
		"rounded":         true,
		"bg-red-100":      errs.Has(name),
	}

	for _, class := range extraClasses {
		classes[class] = true
	}

	return Group{
		If(errs.Has(name), P(Class("pt-5 text-red-500 text-sm italic"), Text(errs.Get(name)))),
		Textarea(classes,
			Name(name),
			Text(form.Get(name)),
			Rows("3"),
		),
	}
}
