package crawler

import (
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"net/http"
	"net/url"
)

const robotsPath = "/robots.txt"

// Get once the robots file. See robotGet for details.
func robotGetter(database *DB, host string, roundTripper http.RoundTripper) func() *robotstxt.File {
	robot := robotstxt.File{}
	todo := true
	return func() *robotstxt.File {
		if todo {
			todo = false
			robot = robotGet(database, host, roundTripper)
		}
		return &robot
	}
}

// RobotGetter load it, or download and store it from the objectDB.
// On error (from cache or when download), use robotstxt.DefaultRobots.
// If cache is more than 24h, dowload it.
func robotGet(database *DB, host string, roundTripper http.RoundTripper) robotstxt.File {
	u := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   robotsPath,
	}
	key := db.NewURLKey(&u)

	// Get from the DB
	if page, _ := database.Object.Get(key); page != nil {
		if page.Robots != nil {
			// check < 24h
			return *page.Robots
		}
		return robotstxt.DefaultRobots
	}

	robots := robotstxt.DefaultRobots
	if buff, _, _ := fetchBytes(&u, roundTripper, 5, 500_1024); buff != nil {
		robots = robotstxt.Parse(buff.Bytes())
		recycler.Recycle(buff)
	}

	database.save(&u, &Page{Robots: &robots})

	return robots
}
