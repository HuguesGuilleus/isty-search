package crawler

import (
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"net/http"
	"net/url"
	"time"
)

type Page struct {
	URL  url.URL
	Time time.Time

	// Content, on of the following filed.
	Error    string
	Html     *htmlnode.Root
	Robots   *robotstxt.File
	Redirect *url.URL
}

// Get once the robots file. See robotGet for details.
func robotGetter(objectDB db.ObjectBD[Page], host string, roundTripper http.RoundTripper) func() robotstxt.File {
	robot := robotstxt.File{}
	todo := true
	return func() robotstxt.File {
		if todo {
			todo = false
			robot = robotGet(objectDB, host, roundTripper)
		}
		return robot
	}
}

// RobotGetter load it, or download and store it from the objectDB.
// On error (from cache or when download), use robotstxt.DefaultRobots.
// If cache is more than 24h, dowload it.
func robotGet(objectDB db.ObjectBD[Page], host string, roundTripper http.RoundTripper) robotstxt.File {
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

	robots := robotstxt.DefaultRobots
	if buff, _, _ := fetchBytes(&u, roundTripper, 5, 500_1024); buff != nil {
		robots = robotstxt.Parse(buff.Bytes())
		recycler.Recycle(buff)
	}

	page.Robots = &robots

	return robots
}
