package crawler

import (
	"context"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"github.com/HuguesGuilleus/isty-search/keys"
	"net/http"
	"net/url"
	"time"
)

const robotsPath = "/robots.txt"
const faviconPath = "/favicon.ico"

// Get once the robots file. See robotGet for details.
func robotGetter(ctx context.Context, db *crawldatabase.Database[Page], scheme, host string, roundTripper http.RoundTripper) func() *robotstxt.File {
	robot := robotstxt.File{}
	todo := true
	return func() *robotstxt.File {
		if todo {
			todo = false
			robot = robotGet(ctx, db, scheme, host, roundTripper)
		}
		return &robot
	}
}

// RobotGetter load it, or download and store it from the objectDB.
// On error (from cache or when download), use robotstxt.DefaultRobots.
// If cache is more than 24h, dowload it.
func robotGet(ctx context.Context, db *crawldatabase.Database[Page], scheme, host string, roundTripper http.RoundTripper) robotstxt.File {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   robotsPath,
	}
	key := keys.NewURL(&u)

	// Get from the DB
	if page, date, _ := db.GetValue(key); page != nil && time.Since(date) < time.Hour*24 {
		if page.Robots != nil {
			return *page.Robots
		}
		return robotstxt.DefaultRobots
	}

	robots := robotstxt.DefaultRobots
	if buff := fetchMultiple(ctx, roundTripper, 500_1024, &u, 5); buff != nil {
		robots = robotstxt.Parse(buff.Bytes())
		common.RecycleBuffer(buff)
	}

	db.SetValue(key, &Page{
		URL:    u,
		Robots: &robots},
		crawldatabase.TypeFileRobots,
	)

	return robots
}
