package crawler

import (
	"errors"
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const robotsMaxRedirect = 5

var robotsMaxRedirectTooManyRedirect = errors.New("Too many redirect (limit=5)")

type Page struct {
	URL  url.URL
	Time time.Time

	// Content, on of the following filed.
	Error  string
	Node   *htmlnode.Node
	Robots *robotstxt.File
}

// getRobotstxt load ir, or download and store it from the objectDB.
// Do not use the context filters.
// On error (from cache or when download), use robotstxt.DefaultRobots.
func getRobotstxt(objectDB db.ObjectBD[Page], host string, roundTripper http.RoundTripper) robotstxt.File {
	u := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   "/robots.txt",
	}
	key := db.NewURLKey(&u)

	// Get from the DB
	if page, _ := objectDB.Get(key); page != nil {
		if page.Robots != nil {
			// check < 24h
			return *page.Robots
		}
		return robotstxt.DefaultRobots
	}

	page := &Page{
		URL:  u,
		Time: time.Now(),
	}
	defer objectDB.Store(key, page)

	// Download it
	client := http.Client{
		Transport: roundTripper,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) > robotsMaxRedirect {
				return robotsMaxRedirectTooManyRedirect
			}
			return nil
		},
		Timeout: robotsMaxRedirect * time.Second,
	}
	response, err := client.Get(u.String())
	if err != nil {
		page.Error = err.Error()
		return robotstxt.DefaultRobots
	}

	// Parse the response
	buff := recycler.Get()
	defer recycler.Recycle(buff)
	defer response.Body.Close()
	if _, err := io.CopyN(buff, response.Body, 500_1024); err != nil && err != io.EOF {
		page.Error = err.Error()
		return robotstxt.DefaultRobots
	}
	robots := robotstxt.Parse(buff.Bytes())

	page.Robots = &robots

	return robots
}
