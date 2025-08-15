package domain

import (
	"strings"

	"github.com/samber/lo"
)

// StopWord is the stopword domain model.
type StopWord struct {
	Word string
}

// StopWordList is a domain model for the list of stop words.
type StopWordList []StopWord

// NewStopWordList creates a new stop word list.
func NewStopWordList(words ...string) StopWordList {
	return lo.Uniq(lo.FilterMap(words, func(word string, _ int) (StopWord, bool) {
		word = strings.ToLower(strings.TrimSpace(word))
		if word == "" {
			return StopWord{}, false
		}

		return StopWord{Word: word}, true
	}))
}
