package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt/test_data"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetRobotstxt(t *testing.T) {
	db := db.OpenObjectBD[Page]("__test_db")
	defer os.RemoveAll("__test_db")

	robots := robotGetter(db, "www.monde-diplomatique.fr", mapRoundTripper{
		"https://www.monde-diplomatique.fr/robots.txt": robotstxtTestData.MondeDiplomatique,
	})()
	assert.Equal(t, robotstxt.Parse(robotstxtTestData.MondeDiplomatique), robots)

	robotsSecond := robotGetter(db, "www.monde-diplomatique.fr", mapRoundTripper{})()
	// The assert package make a difference between empty slice and nil slice,
	// so we test only CrawlDelay (type int).
	assert.Equal(t, robotstxt.Parse(robotstxtTestData.MondeDiplomatique).CrawlDelay, robotsSecond.CrawlDelay)
}
