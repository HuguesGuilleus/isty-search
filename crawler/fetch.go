package crawler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"io"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type fetchContext struct {
	db *crawldatabase.Database[Page]

	// The Hosts, a map to store all urls that will be crawled.
	// The key is generate by createKey().
	hosts      map[string]*host
	hostsMutex sync.Mutex

	// A parent context
	context context.Context
	// Number of current crawl goroutine.
	lenGo int
	// Maximum of crawl goroutine
	maxGo int
	// Done when all crawl goroutine return.
	wg sync.WaitGroup

	filterURL  []func(*url.URL) bool
	filterPage []func(*htmlnode.Root) bool

	roundTripper http.RoundTripper

	// The max size of the html page.
	maxLength int64

	// The min and max CrawlDelay.
	// The used value if determined by the robots.txt.
	// Must: minCrawlDelay < maxCrawlDelay
	minCrawlDelay, maxCrawlDelay time.Duration
}

type host struct {
	scheme   string
	host     string
	urls     []*url.URL
	fetching bool
}

func (ctx *fetchContext) Work() {
	defer ctx.wg.Done()
	for h := ctx.tryChooseWork(nil); h != nil && ctx.context.Err() == nil; h = ctx.tryChooseWork(h) {
		ctx.crawlHost(h)
	}
}

// Crawl (strike and )
func (ctx *fetchContext) crawlHost(h *host) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("[defered error]", err)
			debug.PrintStack()
		}
	}()

	urls, crawDelay := ctx.strikeURLs(h)
	for _, u := range urls {
		ctx.sleep(crawDelay)
		ctx.fetchOne(u)
		if ctx.context.Err() != nil {
			return
		}
	}
}

// Get a host that will be crawled, and free last host if not nil.
func (ctx *fetchContext) tryChooseWork(lastHost *host) *host {
	ctx.hostsMutex.Lock()
	defer ctx.hostsMutex.Unlock()

	if lastHost != nil {
		key := createKey(lastHost.scheme, lastHost.host)
		ctx.hosts[key].fetching = false
		if len(ctx.hosts[key].urls) == 0 {
			delete(ctx.hosts, key)
		}
	}

	for _, h := range ctx.hosts {
		if !h.fetching {
			h.fetching = true
			returnedHost := *h
			h.urls = nil
			return &returnedHost
		}
	}

	ctx.lenGo--
	return nil
}

// Add urls in the URLsDB and in the ctx.host, then lauch if it's possible new crawl goroutine.
func (ctx *fetchContext) addURLs(urls map[crawldatabase.Key]*url.URL) {
	ctx.db.AddURL(urls)
	ctx.planURLs(urls)
}

func (ctx *fetchContext) planURLs(urls map[crawldatabase.Key]*url.URL) {
	ctx.hostsMutex.Lock()
	defer ctx.hostsMutex.Unlock()

	for _, u := range urls {
		key := createKey(u.Scheme, u.Host)
		h := ctx.hosts[key]
		if h == nil {
			h = &host{
				scheme: u.Scheme,
				host:   u.Host,
				urls:   make([]*url.URL, 0),
			}
			ctx.hosts[key] = h
		}
		h.urls = append(h.urls, u)
	}

	max := len(ctx.hosts)
	if max > ctx.maxGo {
		max = ctx.maxGo
	}
	for i := ctx.lenGo; i < max; i++ {
		ctx.wg.Add(1)
		ctx.lenGo++
		go ctx.Work()
	}
}

// Join the scheme and the host with two point.
func createKey(scheme, host string) string { return scheme + ":" + host }

/* FETCHING ONE */

func (ctx *fetchContext) fetchOne(u *url.URL) {
	key := crawldatabase.NewKeyURL(u)

	// Get the body
	body, redirect, errString := fetchBytes(ctx.context, ctx.roundTripper, ctx.maxLength, u)
	if errString {
		ctx.db.SetSimple(key, crawldatabase.TypeErrorNetwork)
		return
	} else if redirect != nil {
		ctx.addURLs(map[crawldatabase.Key]*url.URL{
			crawldatabase.NewKeyURL(redirect): redirect,
		})
		ctx.db.SetRedirect(key, crawldatabase.NewKeyURL(redirect))
		return
	}
	defer common.RecycleBuffer(body)

	// Parse the body
	htmlRoot, err := htmlnode.Parse(body.Bytes())
	if err != nil {
		ctx.db.SetSimple(key, crawldatabase.TypeErrorParsing)
		return
	}

	// Post filter
	if htmlRoot.Meta.NoIndex {
		ctx.db.SetSimple(key, crawldatabase.TypeErrorNoIndex)
		return
	}
	for _, filter := range ctx.filterPage {
		if filter(htmlRoot) {
			ctx.db.SetSimple(key, crawldatabase.TypeErrorFilterPage)
			return
		}
	}

	page := &Page{
		URL:  *u,
		Html: htmlRoot,
	}

	// Get URL
	if !htmlRoot.Meta.NoFollow {
		ctx.addURLs(page.GetURLs())
	}

	// Save it
	ctx.db.SetValue(key, page, crawldatabase.TypeFileHTML)

	return
}

// Strike all url (from same host), and return it with crawDelay.
// - The path is "/robots.txt" or "/favicon.ico"
// - Filtered
// - Blocked by robots.
func (ctx *fetchContext) strikeURLs(h *host) ([]*url.URL, int) {
	robotsGetter := robotGetter(ctx.context, ctx.db, h.scheme, h.host, ctx.roundTripper)
	validURLs := make([]*url.URL, 0, len(h.urls))

urlFor:
	for _, u := range h.urls {
		switch u.Path {
		case robotsPath, faviconPath:
			continue urlFor
		}

		// Context filters
		for _, filter := range ctx.filterURL {
			if filter(u) {
				ctx.db.SetSimple(crawldatabase.NewKeyURL(u), crawldatabase.TypeErrorFilterURL)
				continue urlFor
			}
		}

		// Robots.txt
		if !robotsGetter().Allow(u) {
			ctx.db.SetSimple(crawldatabase.NewKeyURL(u), crawldatabase.TypeErrorRobot)
			continue urlFor
		}

		validURLs = append(validURLs, u)
	}

	if len(validURLs) == 0 {
		return nil, 0
	}

	return validURLs, robotsGetter().CrawlDelay
}

func (ctx *fetchContext) sleep(crawDelay int) {
	delay := time.Duration(crawDelay) * time.Second
	if delay < ctx.minCrawlDelay {
		delay = ctx.minCrawlDelay
	}
	if delay > ctx.maxCrawlDelay {
		delay = ctx.maxCrawlDelay
	}

	timeoutContext, cancel := context.WithTimeout(ctx.context, delay)
	defer cancel()
	<-timeoutContext.Done()
}

// Fetch until the redirection is over maxRedirect.
// Used my robotGet()
func fetchMultiple(ctx context.Context, roundTripper http.RoundTripper, maxLength int64, u *url.URL, maxRedirect int) (buff *bytes.Buffer) {
	for i := 0; i < maxRedirect && u != nil; i++ {
		buff, u, _ = fetchBytes(ctx, roundTripper, maxLength, u)
	}
	return
}

// Fetch the url, and return: the body, the redirect URL or the error.
func fetchBytes(ctx context.Context, roundTripper http.RoundTripper, maxLength int64, u *url.URL) (*bytes.Buffer, *url.URL, bool) {
	if h := u.Host; strings.LastIndex(h, ":") > strings.LastIndex(h, "]") {
		u.Host = strings.TrimSuffix(h, ":")
	}

	request := http.Request{
		Method:     http.MethodGet,
		URL:        u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       u.Host,
	}
	requestContext, cancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer cancel()
	request.WithContext(requestContext)
	response, err := roundTripper.RoundTrip(&request)

	if err != nil {
		return nil, nil, true
	}
	defer response.Body.Close()

	if code := response.StatusCode / 100; code == 3 {
		return getLocation(u, response)
	} else if code != 2 {
		return nil, nil, true
	}

	buff := common.GetBuffer()
	if l := response.ContentLength; l > 0 && l < maxLength {
		buff.Grow(int(l))
	} else {
		buff.Grow(int(maxLength))
	}
	if _, err := buff.ReadFrom(io.LimitReader(response.Body, maxLength)); err != nil {
		common.RecycleBuffer(buff)
		return nil, nil, true
	}

	return buff, nil, false
}

// Get the location from response headers.
// The bytes buffer is allways nil.
func getLocation(u *url.URL, response *http.Response) (*bytes.Buffer, *url.URL, bool) {
	redirectString := response.Header.Get("Location")
	if redirectString == "" {
		return nil, nil, true
	}
	redirect, err := u.Parse(redirectString)
	if err != nil {
		return nil, nil, true
	}
	cleanURL(redirect)
	if redirect.String() == u.String() {
		return nil, nil, true
	}
	return nil, redirect, false
}
