package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"log"
	"net/url"
	"time"
)

type fetchContext struct {
	host      string
	urls      []*url.URL
	db        *DB
	outputURL chan<- []*url.URL

	filterURL  []func(*url.URL) string
	filterPage []func(*htmlnode.Node) string

	// Function to process all page.
	process []func(*htmlnode.Node)

	// The min and max CrawlDelay.
	// The used value if determined by the robots.txt.
	// Must: minCrawlDelay < maxCrawlDelay
	minCrawlDelay, maxCrawlDelay time.Duration

	log *log.Logger
}

func fetch(context fetchContext) {}
