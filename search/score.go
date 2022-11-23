package search

import (
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"sort"
)

type Score struct {
	Key  db.Key
	Rank float32
}

// Sort the score by the rank.
func SortScores(scores []Score) {
	sort.Slice(scores, func(i, j int) bool { return scores[i].Rank > scores[j].Rank })
}
