package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt/testdata"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetRobotstxt(t *testing.T) {
	defer os.RemoveAll("__test_db")

	database := &DB{
		KeyValueDB: db.OpenKeyValueDB[Page]("__test_db"),
	}

	robots := *robotGetter(database, "https", "www.monde-diplomatique.fr", mapRoundTripper{
		"https://www.monde-diplomatique.fr/robots.txt": robotstxttestdata.MondeDiplomatique,
	})()
	assert.Equal(t, robotstxt.Parse(robotstxttestdata.MondeDiplomatique), robots)

	robotsSecond := *robotGetter(database, "https", "www.monde-diplomatique.fr", mapRoundTripper{})()
	// The assert package make a difference between empty slice and nil slice,
	// so we test only CrawlDelay (type int).
	assert.Equal(t, robotstxt.Parse(robotstxttestdata.MondeDiplomatique).CrawlDelay, robotsSecond.CrawlDelay)
}
