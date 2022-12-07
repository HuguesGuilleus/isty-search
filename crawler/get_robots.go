package crawler

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"net/http"
	"net/url"
	"time"
)

const robotsPath = "/robots.txt"

// Get once the robots file. See robotGet for details.
func robotGetter(database *DB, scheme, host string, roundTripper http.RoundTripper) func() *robotstxt.File {
	robot := robotstxt.File{}
	todo := true
	return func() *robotstxt.File {
		if todo {
			todo = false
			robot = robotGet(database, scheme, host, roundTripper)
		}
		return &robot
	}
}

// RobotGetter load it, or download and store it from the objectDB.
// On error (from cache or when download), use robotstxt.DefaultRobots.
// If cache is more than 24h, dowload it.
func robotGet(database *DB, scheme, host string, roundTripper http.RoundTripper) robotstxt.File {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   robotsPath,
	}
	key := db.NewURLKey(&u)

	// Get from the DB
	if page, _ := database.KeyValueDB.Get(key); page != nil && time.Since(page.Time) < time.Hour*24 {
		if page.Robots != nil {
			return *page.Robots
		}
		return robotstxt.DefaultRobots
	}

	robots := robotstxt.DefaultRobots
	if buff, _, _ := fetchBytes(&u, roundTripper, 5, 500_1024); buff != nil {
		robots = robotstxt.Parse(buff.Bytes())
		common.RecycleBuffer(buff)
	}

	database.save(&u, &Page{Robots: &robots})

	return robots
}
