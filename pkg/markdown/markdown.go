package markdown

import (
	"strings"

	"github.com/yuin/goldmark"
)

// Convert a markdown source to HTML.
func Convert(source string) (string, error) {
	var buf strings.Builder
	if err := goldmark.Convert([]byte(source), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}
