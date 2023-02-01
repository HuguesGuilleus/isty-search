package index

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"sort"
)

// Simple crawler.Processor that count the vocabulary occure.
type CounterVocab map[string]int

func (counter CounterVocab) Process(page *crawler.Page) {
	page.Html.Body.Visit(func(node htmlnode.Node) {
		for _, word := range getVocab(node.Text) {
			counter[word]++
		}
	})
}

// Total numbers of vocabulary all document.
// Must be call after all VocabCounter.Process()
func (counter CounterVocab) Sum() (sum int) {
	for _, nb := range counter {
		sum += nb
	}
	return
}

type WordFrequency struct {
	Word  string
	Count int
}

// Get the word frequency list sorted in reverse order.
func (counter CounterVocab) Frequency() (list []WordFrequency) {
	list = make([]WordFrequency, 0, len(counter))
	for word, count := range counter {
		list = append(list, WordFrequency{word, count})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count == list[j].Count {
			return list[i].Word < list[j].Word
		}
		return list[i].Count > list[j].Count
	})

	return
}
