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
	return lo.FilterMap(words, func(item string, _ int) (StopWord, bool) {
		item = strings.TrimSpace(item)
		if item == "" {
			return StopWord{}, false
		}

		return StopWord{Word: item}, true
	})
}
