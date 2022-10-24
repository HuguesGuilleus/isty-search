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

	robots := getRobotstxt(db, "www.monde-diplomatique.fr", mapRoundTripper{
		"https://www.monde-diplomatique.fr/robots.txt": robotstxtTestData.MondeDiplomatique,
	})
	assert.Equal(t, robotstxt.Parse(robotstxtTestData.MondeDiplomatique), robots)

	robotsSecod := getRobotstxt(db, "www.monde-diplomatique.fr", mapRoundTripper{})
	// We test only CrawlDelay because the rules can contain empty slice or
	// nil slice and for the assert packeg is different.
	assert.Equal(t, robotstxt.Parse(robotstxtTestData.MondeDiplomatique).CrawlDelay, robotsSecod.CrawlDelay)
}
