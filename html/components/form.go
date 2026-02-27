package components

import (
	"strconv"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ButtonElement is a button.
func ButtonElement(text string, children ...Node) Node {
	nodes := []Node{
		buttonClasses(),
		Type("button"),
		Text(text),
	}
	nodes = append(nodes, children...)

	// TODO: Group can't use attrs?
	return Button(
		nodes...,
	)
}

// SubmitButtonElement is a submit button.
func SubmitButtonElement(text string, children ...Node) Node {
	nodes := []Node{
		buttonClasses(),
		Type("submit"),
		Text(text),
	}
	nodes = append(nodes, children...)

	// TODO: Group can't use attrs?
	return Button(
		nodes...,
	)
}

// InputElement is an input element.
// TODO: struct arguments
func InputElement(name, typ, placeholder string, value, err string, required, autocomplete bool) Node {
	return withErrors(err,
		Input(baseInputClasses(err != ""),
			Name(name),
			Type(typ),
			If(required, Placeholder(placeholder+"*")),
			If(!required, Placeholder(placeholder)),
			Value(value),
			If(required, Required()),
			If(!autocomplete, AutoComplete("off")),
		),
	)
}

// TextareaElement is a textarea element.
func TextareaElement(name string, value, err string, rows uint64, required, autocomplete bool) Node {
	return withErrors(err,
		Textarea(baseInputClasses(err != ""),
			Name(name),
			Text(value),
			Rows(strconv.FormatUint(rows, 10)),
			If(required, Required()),
			If(!autocomplete, AutoComplete("off")),
		),
	)
}

// DateTimeLocalInput is a datetime-local element.
func DateTimeLocalInput(name string, value, err string, required, autocomplete bool) Node {
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
