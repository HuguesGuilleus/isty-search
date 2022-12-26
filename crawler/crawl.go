package crawler

import (
	"context"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"golang.org/x/exp/slog"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	// The database, DB or the root if DB is nil.
	DB crawldatabase.Database[Page]

	// Root URL to begin to read
	Input []*url.URL

	// Filter by URL or by the page (for exemple by the language).
	// Return true to strike the page.
	// The file: "/robots.txt" and "/favicon.ico" are not tested.
	FilterURL  []func(*url.URL) bool
	FilterPage []func(*htmlnode.Root) bool

	// The max size of the html page.
	// 15M for Google https://developers.google.com/search/docs/crawling-indexing/googlebot#how-googlebot-accesses-your-site
	MaxLength int64

	// Maximum of crawl goroutine
	MaxGo int

	// The min and max CrawlDelay.
	// The used value if determined by the robots.txt.
	// Must: minCrawlDelay < maxCrawlDelay
	MinCrawlDelay, MaxCrawlDelay time.Duration

	Logger *slog.Logger

	// Use to fetch all HTTP ressource.
	RoundTripper http.RoundTripper
}

func Crawl(mainContext context.Context, config Config) error {
	end := make(chan struct{})
	fetchContext := &fetchContext{
		db:            config.DB,
		hosts:         make(map[string]*host),
		end:           end,
		maxGo:         config.MaxGo,
		filterURL:     config.FilterURL,
		filterPage:    config.FilterPage,
		roundTripper:  newlogRoundTripper(config.RoundTripper, config.Logger),
		maxLength:     config.MaxLength,
		minCrawlDelay: config.MinCrawlDelay,
		maxCrawlDelay: config.MaxCrawlDelay,
	}
	defer fetchContext.wg.Wait()

	if len(config.Input) == 0 {
		return nil
	}

	dbUrls := make(map[crawldatabase.Key]*url.URL, len(config.Input))
	for _, u := range config.Input {
		dbUrls[crawldatabase.NewKeyURL(u)] = u
	}
	fetchContext.db.AddURL(dbUrls)

	urls := make(map[crawldatabase.Key]*url.URL, len(config.Input))
	for _, u := range config.Input {
		urls[crawldatabase.NewKeyURL(u)] = u
	}
	fetchContext.planURLs(urls)

	select {
	case <-end:
	case <-mainContext.Done():
		fetchContext.close = true
	}

	return nil
}
