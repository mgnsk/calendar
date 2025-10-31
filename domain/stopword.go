package domain

import (
	"strings"

	"github.com/samber/lo"
)

// StopWordList is a domain model for the list of stop words.
// Stop words are lowercase and deduplicated.
type StopWordList []string

// NewStopWordList creates a new stop word list.
func NewStopWordList(words ...string) StopWordList {
	return lo.Uniq(lo.FilterMap(words, func(word string, _ int) (string, bool) {
		word = strings.ToLower(strings.TrimSpace(word))
		if word == "" {
			return "", false
		}

		return word, true
	}))
}
