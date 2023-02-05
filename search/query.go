package search

import (
	"github.com/HuguesGuilleus/isty-search/index"
	"github.com/HuguesGuilleus/isty-search/keys"
)

type Query struct {
	// The normalized word from the user input.
	Word string
	// The key of the word.
	Key keys.Key
	// Number of page with this article
	Count int
}

func parse(q string) []Query {
	splited := index.GetVocab(q)
	queries := make([]Query, len(splited))
	for i, s := range splited {
		queries[i] = Query{
			Word: s,
			Key:  keys.NewString(s),
		}
	}
	return queries
}
