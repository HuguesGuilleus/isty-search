package search

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"strings"
)

type VocabAdvanced map[crawldatabase.Key][]struct {
	page  crawldatabase.Key
	count int
}

func (advanced VocabAdvanced) Process(page *crawler.Page) {
	counter := make(map[string]int)

	page.Html.Body.Visit(func(node htmlnode.Node) {
		text := node.Text
		text = strings.TrimSpace(text)
		text = strings.ToLower(text)
		if text == "" {
			return
		}
		for _, word := range strings.FieldsFunc(text, splitWords) {
			counter[word]++
		}
	})

	key := crawldatabase.NewKeyURL(&page.URL)
	for word, count := range counter {
		kw := crawldatabase.NewKeyString(word)
		advanced[kw] = append(advanced[kw], struct {
			page  crawldatabase.Key
			count int
		}{key, count})
	}
}
