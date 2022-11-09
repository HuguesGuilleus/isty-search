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
	defer os.RemoveAll("__test_db")

	database := &DB{
		Object:    db.OpenObjectBD[Page]("__test_db"),
		Existence: db.OpenExistenceMap(),
	}

	robots := *robotGetter(database, "https", "www.monde-diplomatique.fr", mapRoundTripper{
		"https://www.monde-diplomatique.fr/robots.txt": robotstxtTestData.MondeDiplomatique,
	})()
	assert.Equal(t, robotstxt.Parse(robotstxtTestData.MondeDiplomatique), robots)

	robotsSecond := *robotGetter(database, "https", "www.monde-diplomatique.fr", mapRoundTripper{})()
	// The assert package make a difference between empty slice and nil slice,
	// so we test only CrawlDelay (type int).
	assert.Equal(t, robotstxt.Parse(robotstxtTestData.MondeDiplomatique).CrawlDelay, robotsSecond.CrawlDelay)
}
