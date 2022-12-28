package crawler

import (
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt/testdata"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetRobotstxt(t *testing.T) {
	_, db, _ := crawldatabase.OpenMemory[Page](nil, "", false)

	robots := *robotGetter(db, "https", "www.monde-diplomatique.fr", mapRoundTripper{
		"https://www.monde-diplomatique.fr/robots.txt": robotstxttestdata.MondeDiplomatique,
	})()
	assert.Equal(t, robotstxt.Parse(robotstxttestdata.MondeDiplomatique), robots)

	robotsSecond := *robotGetter(db, "https", "www.monde-diplomatique.fr", mapRoundTripper{})()
	// The assert package make a difference between empty slice and nil slice,
	// so we test only CrawlDelay (type int).
	assert.Equal(t, robotstxt.Parse(robotstxttestdata.MondeDiplomatique).CrawlDelay, robotsSecond.CrawlDelay)
}
