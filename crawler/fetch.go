package crawler

import (
	"bytes"
	"errors"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime/debug"
	"sync"
	"time"
)

var tooRedirect = errors.New("Too redirect")

type fetchContext struct {
	db *crawldatabase.Database[Page]

	// The Hosts, a map to store all urls that will be crawled.
	// The key is generate by createKey().
	hosts      map[string]*host
	hostsMutex sync.Mutex

	// Channel to signal crawl end.
	end chan<- struct{}
	// Ask to close.
	close bool
	// Done when all crawl goroutine return.
	wg sync.WaitGroup

	// Number of current crawl goroutine.
	lenGo int
	// Maximum of crawl goroutine
	maxGo int

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
	defer func() {
		if err := recover(); err != nil {
			log.Println("[defered error]", err)
			debug.PrintStack()
		}
	}()

	for h := ctx.tryChooseWork(nil); h != nil; h = ctx.tryChooseWork(h) {
		urls, crawDelay := ctx.strikeURLs(h.scheme, h.host, h.urls)
		for _, u := range urls {
			fetchOne(ctx, u)
			ctx.sleep(crawDelay)
			if ctx.close {
				return
			}
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

	fetching := false
	for _, h := range ctx.hosts {
		if h.fetching {
			fetching = true
		} else {
			h.fetching = true
			returnedHost := *h
			h.urls = nil
			return &returnedHost
		}
	}

	if !fetching {
		ctx.end <- struct{}{}
		close(ctx.end)
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

func fetchOne(ctx *fetchContext, u *url.URL) {
	key := crawldatabase.NewKeyURL(u)

	// Get the body
	body, redirect, errString := fetchBytes(u, ctx.roundTripper, 1, ctx.maxLength)
	if errString != "" {
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
// - The path is "/robots.txt"
// - Filtered
// - Blocked by robots.
func (ctx *fetchContext) strikeURLs(scheme, host string, urls []*url.URL) ([]*url.URL, int) {
	robotsGetter := robotGetter(ctx.db, scheme, host, ctx.roundTripper)
	validURLs := make([]*url.URL, 0, len(urls))

urlFor:
	for _, u := range urls {
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
	time.Sleep(delay)
}

// Fetch the url, and return: the body, the redirect URL or the error.
func fetchBytes(u *url.URL, roundTripper http.RoundTripper, maxRedirect int, maxBody int64) (*bytes.Buffer, *url.URL, string) {
	client := http.Client{
		Transport: roundTripper,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) > maxRedirect {
				return tooRedirect
			}
			return nil
		},
		Timeout: time.Duration(maxRedirect) * time.Second,
	}

	response, err := client.Get(u.String())
	if err != nil {
		if errors.Is(err, tooRedirect) {
			return getLocation(u, response)
		}
		return nil, nil, err.Error()
	}
	defer response.Body.Close()

	if code := response.StatusCode / 100; code == 3 {
		return getLocation(u, response)
	} else if code != 2 {
		return nil, nil, "http error:" + response.Status
	}

	buff := common.GetBuffer()
	if l := response.ContentLength; l > 0 && l < maxBody {
		buff.Grow(int(l))
	} else {
		buff.Grow(int(maxBody))
	}
	if _, err := buff.ReadFrom(io.LimitReader(response.Body, maxBody)); err != nil {
		common.RecycleBuffer(buff)
		return nil, nil, err.Error()
	}

	return buff, nil, ""
}

// Get the location from response headers.
// The bytes buffer is allways nil.
func getLocation(u *url.URL, response *http.Response) (*bytes.Buffer, *url.URL, string) {
	redirectString := response.Header.Get("Location")
	if redirectString == "" {
		return nil, nil, "redirect without URL"
	}
	redirect, err := u.Parse(redirectString)
	if err != nil {
		return nil, nil, "redirect with wrong syntax"
	}
	redirect.RawQuery = redirect.Query().Encode()
	if redirect.String() == u.String() {
		return nil, nil, "redirect to the same url"
	}
	return nil, redirect, ""
}
