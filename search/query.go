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
	if len(splited) == 0 {
		return nil
	}
	queries := make([]Query, len(splited))
	for i, s := range splited {
		queries[i] = Query{
			Word: s,
			Key:  keys.NewString(s),
		}
	}
	return queries
}

// Merge common element of the two slice into a.
// Return the a truncated with common element.
// The tow slice must be sorted by key.
func commonKeyFloat32s(a, b []index.KeyFloat32) []index.KeyFloat32 {
	writeIndex := 0
	ai := 0
	bi := 0
	for ai < len(a) && bi < len(b) {
		if a[ai].Key == b[bi].Key {
			a[writeIndex] = a[ai]
			writeIndex++
			ai++
			bi++
		} else if a[ai].Key.Less(&b[bi].Key) {
			ai++
		} else {
			bi++
		}
	}
	return a[:writeIndex]
}
