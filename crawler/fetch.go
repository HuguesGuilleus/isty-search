package crawler

import (
	"context"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type fetchContext struct {
	db        *DB
	outputURL chan<- []*url.URL

	filterURL  []func(*url.URL) string
	filterPage []func(*htmlnode.Root) string

	roundTripper http.RoundTripper
	// The max size of the html page.
	maxLength int64

	// Function to process all page.
	process []func(*Page)

	// The min and max CrawlDelay.
	// The used value if determined by the robots.txt.
	// Must: minCrawlDelay < maxCrawlDelay
	minCrawlDelay, maxCrawlDelay time.Duration

	logOutput io.Writer
	logMutex  sync.Mutex
}

func fetchList(context *fetchContext, host string, urls []*url.URL) {
	urls = context.db.Existence.Filter(urls)

	if len(urls) == 0 {
		return
	}

	robot := robotGet(context.db.Object, host, context.roundTripper)

	for _, u := range urls {
		fetchOne(context, robot, u)
		context.sleep(robot)
	}
}

func fetchOne(ctx *fetchContext, robot robotstxt.File, u *url.URL) {
	key := db.NewURLKey(u)
	defer ctx.db.Existence.Add(key)

	ban := func(reason string) {
		ctx.log("filter", u, reason)
		ctx.db.Ban.Add(u)
		ctx.db.Object.Store(key, &Page{
			URL:   *u,
			Time:  now(),
			Error: reason,
		})
		return
	}

	// URL strike
	if !robot.Allow(u) {
		ban("robots.txt")
		return
	}
	for _, filter := range ctx.filterURL {
		if strike := filter(u); strike != "" {
			ban(strike)
			return
		}
	}

	// get the body
	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer ctxCancel()
	request, err := http.NewRequestWithContext(ctxTimeout, http.MethodGet, u.String(), nil)
	if err != nil {
		ban(err.Error())
		return
	}
	response, err := ctx.roundTripper.RoundTrip(request)
	if err != nil {
		ban(err.Error())
		return
	} else if response.StatusCode/100 != 2 {
		ban("http.noOK")
		return
	}
	buff := recycler.Get()
	defer recycler.Recycle(buff)
	defer response.Body.Close()
	if _, err := buff.ReadFrom(&io.LimitedReader{
		R: response.Body,
		N: ctx.maxLength,
	}); err != nil && err != io.EOF {
		ban(err.Error())
		return
	}
	htmlRoot, err := htmlnode.Parse(buff.Bytes())

	// Post filter
	if htmlRoot.Meta.NoIndex {
		ban("noindex")
		return
	}
	for _, filter := range ctx.filterPage {
		if strike := filter(htmlRoot); strike != "" {
			ban(strike)
			return
		}
	}

	// Get URL
	if !htmlRoot.Meta.NoFollow {
		ctx.outputURL <- htmlRoot.GetURL(u)
	}

	// Save it
	ctx.log("store", u)
	page := &Page{
		URL:  *u,
		Time: now(),
		Html: htmlRoot,
	}
	ctx.db.Object.Store(key, page)
	for _, process := range ctx.process {
		process(page)
	}
}

func now() time.Time {
	return time.Now().UTC()
}

func (ctx *fetchContext) log(op string, u *url.URL, args ...any) {
	buff := recycler.Get()
	defer recycler.Recycle(buff)

	fmt.Fprintf(buff, "%s [%s] <%s>", time.Now().Format("2006-01-02 15:04:05"), op, u)
	for _, arg := range args {
		fmt.Fprintf(buff, " %v", arg)
	}
	buff.WriteByte('\n')

	ctx.logMutex.Lock()
	defer ctx.logMutex.Unlock()
	buff.WriteTo(ctx.logOutput)
}

func (ctx *fetchContext) sleep(robots robotstxt.File) {
	delay := time.Duration(robots.CrawlDelay) * time.Second
	if delay < ctx.minCrawlDelay {
		delay = ctx.minCrawlDelay
	}
	if delay > ctx.maxCrawlDelay {
		delay = ctx.maxCrawlDelay
	}
	time.Sleep(delay)
}
