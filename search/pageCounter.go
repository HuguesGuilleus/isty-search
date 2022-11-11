package search

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
)

// Simple crawler.Processor that count the number of page processed.
type PageCounter int

func (counter *PageCounter) Process(page *crawler.Page) {
	(*counter)++
}
