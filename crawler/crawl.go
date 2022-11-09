package crawler

import (
	"context"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"io"
	"net/url"
	"time"
)

type Config struct {
	// The database, DB or the root if DB is nil.
	DBRoot string
	DB     *DB

	// Root URL to begin to read
	Input []string

	// Filter by URL of the page (for exemple by the language).
	// Empty string return signify no error.
	FilterURL  []func(*url.URL) string
	FilterPage []func(*htmlnode.Root) string

	// Function to process all page.
	Process []func(*Page)

	// The max size of the html page.
	// 15M for Google https://developers.google.com/search/docs/crawling-indexing/googlebot#how-googlebot-accesses-your-site
	MaxLength int64

	// Maximum of crawl goroutine
	MaxGo int

	// The min and max CrawlDelay.
	// The used value if determined by the robots.txt.
	// Must: minCrawlDelay < maxCrawlDelay
	MinCrawlDelay, MaxCrawlDelay time.Duration

	// log Output.
	// No log for nil value.
	LogOutput io.Writer
}

func Crawl(mainContext context.Context, config Config) error {
	db, foundURLs, err := OpenDB(config.DBRoot)
	if err != nil {
		return fmt.Errorf("Open Composite DB (%q): %w", config.DBRoot, err)
	}
	defer db.Close()

	input := make([]*url.URL, len(config.Input))
	for i, s := range config.Input {
		u, err := url.Parse(s)
		if err != nil {
			return fmt.Errorf("Parse URL %q: %w", s, err)
		}
		input[i] = u
	}

	end := make(chan struct{})
	fetchContext := &fetchContext{
		db:            db,
		hosts:         make(map[string]*host),
		end:           end,
		maxGo:         config.MaxGo,
		filterURL:     config.FilterURL,
		filterPage:    config.FilterPage,
		roundTripper:  newlogRoundTripper(nil, config.LogOutput),
		maxLength:     config.MaxLength,
		process:       config.Process,
		minCrawlDelay: config.MinCrawlDelay,
		maxCrawlDelay: config.MaxCrawlDelay,
	}
	defer fetchContext.wg.Wait()

	fetchContext.loadURLS(foundURLs)
	fetchContext.addURLs(input)

	select {
	case <-end:
	case <-mainContext.Done():
		fetchContext.close = true
	}

	return nil
}
