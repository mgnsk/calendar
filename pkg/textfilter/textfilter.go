package textfilter

import (
	"strconv"
	"strings"
	"unicode"
)

// EnsureQuoted ensures the string is quoted.
func EnsureQuoted(s string) string {
	s = strings.Trim(s, quotes)
	return strconv.Quote(s)
}

// PrepareFTSSearchStrings prepares an SQLITE FTS search string of quoted words.
func PrepareFTSSearchStrings(s string) (quoted []string) {
	fields := splitString(s)

	for _, field := range fields {
		switch field {
		case "AND", "OR", "NOT":
			quoted = append(quoted, field)

		default:
			quoted = append(quoted, EnsureQuoted(field))
		}
	}

	return quoted
}

// splitString splits a string by one or more runs of whitespace while
// attempting to keep the most common bases of quote usage.
func splitString(s string) []string {
	quoted := false
	return strings.FieldsFunc(s, func(r rune) bool {
		if unicode.In(r, unicode.Quotation_Mark) {
			quoted = !quoted
		}
		return !quoted && unicode.IsSpace(r)
	})
}

// GetTags returns all lowercased words.
func GetTags(s string) []string {
	var tags []string

	// Replace all non-letter and non-number characters with spaces.
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return ' '
	}, s)

	for word := range strings.FieldsSeq(s) {
		if len(word) >= 3 {
			word = strings.ToLower(word)
			tags = append(tags, word)
		}
	}

	return tags
}

const quotes = "\"'`"
