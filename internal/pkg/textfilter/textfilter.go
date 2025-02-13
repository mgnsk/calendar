package textfilter

import (
	"strconv"
	"strings"
	"unicode"
)

// EnsureQuoted ensures the string is quoted.
func EnsureQuoted(s string) string {
	if unquoted, err := strconv.Unquote(s); err == nil {
		// String was quoted, quote again.
		return strconv.Quote(unquoted)
	}
	// Not quoted, quote string.
	return strconv.Quote(s)
}

// PrepareFTSSearchString prepares an SQLITE FTS search string of quoted words.
func PrepareFTSSearchString(s string) string {
	fields := SplitString(s)
	quoted := make([]string, 0, len(fields))

	for _, field := range fields {
		quoted = append(quoted, EnsureQuoted(field))
	}

	s = strings.Join(quoted, " ")

	return s
}

// Clean the string, remove any unwanted characters.
func Clean(s string) string {
	return strings.Map(func(r rune) rune {
		if r == unicode.ReplacementChar {
			return -1
		}
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}, s)
}

// SplitString splits a string by one or more runs of whitespace while
// attempting to keep the most common bases of quote usage.
func SplitString(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if unicode.In(r, unicode.Quotation_Mark) {
			quoted = !quoted
		}
		return !quoted && unicode.IsSpace(r)
	})
}
