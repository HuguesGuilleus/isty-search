package search

import (
	"strconv"
	"testing"

	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
)

func TestDBSearch(t *testing.T) {
	pageResults := make([]PageResult, 0)
	for i := 10; i < 20; i++ {
		istr := strconv.Itoa(i)
		pageResults = append(pageResults, PageResult{
			Key:         keys.Key{byte(i)},
			URL:         *common.ParseURL("https://exemple.org/page-" + istr),
			Title:       "page:" + istr,
			Description: "desc:" + istr,
		})
	}

	result, err := FakeDB().Search(" \t WORD\n", 1)
	assert.NoError(t, err)
	assert.Equal(t, &Result{
		Queries: []Query{Query{
			Word:  "word",
			Key:   keys.NewString("word"),
			Count: 51,
		}},
		Results:         pageResults,
		NumberOfResults: 51,
		NumberOfChunck:  6,
	}, result)
}
