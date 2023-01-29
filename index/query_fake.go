package index

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"time"
)

type fakeQuerier struct{}

// A simple querier
func FakeQuerier() Querier { return fakeQuerier{} }

func (_ fakeQuerier) QueryText(query string) ([]string, []PageResult) {
	result := PageResult{
		Title:            "Titre 1",
		Description:      "Lorem ipsum dolor sit amet, consectetur adipisicing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
		LastModification: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
		URL:              *common.ParseURL("https://www.isty.uvsq.fr/"),
	}

	return getVocab(query), []PageResult{
		result, result, result, result,
		result, result, result, result,
		result, result, result, result,
		result, result, result, result,
		result, result, result, result,
	}
}
