package search

import (
	"sort"

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

func search(queryString string, reverseIndex map[keys.Key][]index.KeyFloat32) ([]Query, []index.KeyFloat32) {
	queries := parse(queryString)
	if len(queries) == 0 {
		return nil, nil
	}

	pages := cloneKeyFloat32s(reverseIndex[queries[0].Key])
	for i, query := range queries {
		queryPages := reverseIndex[query.Key]
		queries[i].Count = len(queryPages)
		pages = commonKeyFloat32s(pages, queryPages)
	}

	return queries, pages
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

func cloneKeyFloat32s(src []index.KeyFloat32) []index.KeyFloat32 {
	if len(src) == 0 {
		return nil
	}
	new := make([]index.KeyFloat32, len(src))
	copy(new, src)
	return new
}

// Merge common elements of the two slice into a.
// Return a truncated with common elements.
// The tow slices must be sorted by key.
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

// Add a global score to the pages F32, and sort in reverse order by F32 the pages.
func score(pages []index.KeyFloat32, globalScore map[keys.Key]float32) {
	for i, p := range pages {
		pages[i].F32 += globalScore[p.Key]
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].F32 > pages[j].F32
	})
}
