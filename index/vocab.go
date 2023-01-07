package search

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"strings"
	"unicode"
)

type VocabCounter map[string]int

func (counter VocabCounter) Process(page *crawler.Page) {
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
}

// Total numbers of vocabulary all document.
// Must be call after VocabCounter.Process()
func (counter VocabCounter) TotalWords() (sum int) {
	for _, nb := range counter {
		sum += nb
	}
	return
}

func splitWords(c rune) bool { return c != '-' && !unicode.IsLetter(c) && !unicode.IsNumber(c) }
