package index

import (
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"sort"
)

type Score struct {
	Key  crawldatabase.Key
	Rank float32
}

// Sort the score by the rank.
func SortScores(scores []Score) {
	sort.Slice(scores, func(i, j int) bool { return scores[i].Rank > scores[j].Rank })
}
