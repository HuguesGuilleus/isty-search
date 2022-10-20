package crawler

import (
	"context"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"io"
	"net/url"
	"time"
)

type Config struct {
	Context context.Context

	// The database, DB or the root if Db is nil.
	DBRoot string
	DB     *DB

	// Root URL to begin to read
	Input []string

	// Filter by URL of the page (for exemple by the language).
	// Empty string return signify no error.
	FilterURL  []func(*url.URL) string
	FilterPage []func(*htmlnode.Node) string

	// Function to process all page.
	Process []func(*htmlnode.Node)

	// The min and max CrawlDelay.
	// The used value if determined by the robots.txt.
	// Must: minCrawlDelay < maxCrawlDelay
	minCrawlDelay, maxCrawlDelay time.Duration

	// log Output.
	// No log for nil value.
	LogOutput io.Writer
}

func Crawle(config Config) error {
	return nil
}
