package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestGetStatistics(t *testing.T) {
	assert.Equal(t, Statistics{}, getStatistics(map[keys.Key]metavalue{}))
	assert.Equal(t, Statistics{
		Count: [256]int{
			TypeKnow:     1,
			TypeRedirect: 1,

			TypeFileRobots:  1,
			TypeFileHTML:    1,
			TypeFileRSS:     1,
			TypeFileSitemap: 1,
			TypeFileFavicon: 1,

			TypeErrorNetwork:    1,
			TypeErrorParsing:    1,
			TypeErrorFilterURL:  1,
			TypeErrorFilterPage: 1,
		},
		Total:      11,
		TotalFile:  5,
		TotalError: 4,

		FileSize: [TypeError]int64{
			TypeFileRobots:  2,
			TypeFileHTML:    3,
			TypeFileRSS:     4,
			TypeFileSitemap: 5,
			TypeFileFavicon: 6,
		},

		TotalFileSize: 20,
	}, getStatistics(map[keys.Key]metavalue{
		keys.NewString("key0"): metavalue{Type: TypeKnow},
		keys.NewString("key1"): metavalue{Type: TypeRedirect},

		keys.NewString("key2"): metavalue{Type: TypeFileRobots, Length: 2},
		keys.NewString("key3"): metavalue{Type: TypeFileHTML, Length: 3},
		keys.NewString("key4"): metavalue{Type: TypeFileRSS, Length: 4},
		keys.NewString("key5"): metavalue{Type: TypeFileSitemap, Length: 5},
		keys.NewString("key6"): metavalue{Type: TypeFileFavicon, Length: 6},

		keys.NewString("key7"):  metavalue{Type: TypeErrorNetwork},
		keys.NewString("key8"):  metavalue{Type: TypeErrorParsing},
		keys.NewString("key9"):  metavalue{Type: TypeErrorFilterURL},
		keys.NewString("key10"): metavalue{Type: TypeErrorFilterPage},
	}))
}

func TestStatisticsLog(t *testing.T) {
	stats := Statistics{
		Count: [256]int{
			TypeKnow:            1,
			TypeRedirect:        1,
			TypeFileRobots:      1,
			TypeFileHTML:        1,
			TypeFileRSS:         1,
			TypeFileSitemap:     1,
			TypeFileFavicon:     1,
			TypeErrorNetwork:    1,
			TypeErrorParsing:    1,
			TypeErrorFilterURL:  1,
			TypeErrorFilterPage: 1,
		},
		Total:      11,
		TotalFile:  5,
		TotalError: 4,
		FileSize: [TypeError]int64{
			TypeFileRobots:  2,
			TypeFileHTML:    3,
			TypeFileRSS:     4,
			TypeFileSitemap: 5,
			TypeFileFavicon: 6,
		},
		TotalFileSize: 20,
	}

	records, handler := sloghandlers.NewHandlerRecords(slog.DebugLevel)
	stats.Log(slog.New(handler))
	assert.Equal(t, []string{
		"INFO [db.stats.total] count.all=+011 count.know=+001 count.redirect=+001 count.file=+005 count.error=+004 size=+020",
	}, *records)

	records, handler = sloghandlers.NewHandlerRecords(slog.DebugLevel)
	stats.LogAll(slog.New(handler))
	assert.Equal(t, []string{
		"INFO [db.stats.total] count.all=+011 count.know=+001 count.redirect=+001 count.file=+005 count.error=+004 size=+020",
		"INFO [db.stats.count] count=+001 percent=+009 type=know",
		"INFO [db.stats.count] count=+001 percent=+009 type=redirect",
		"INFO [db.stats.count] count=+001 percent=+009 type=fileRobots",
		"INFO [db.stats.count] count=+001 percent=+009 type=fileHTML",
		"INFO [db.stats.count] count=+001 percent=+009 type=fileRSS",
		"INFO [db.stats.count] count=+001 percent=+009 type=fileSitemap",
		"INFO [db.stats.count] count=+001 percent=+009 type=fileFavicon",
		"INFO [db.stats.count] count=+001 percent=+009 type=errorNetwork",
		"INFO [db.stats.count] count=+001 percent=+009 type=errorParsing",
		"INFO [db.stats.count] count=+001 percent=+009 type=errorFilterURL",
		"INFO [db.stats.count] count=+001 percent=+009 type=errorFilterPage",
		"INFO [db.stats.size] total=+020",
		"INFO [db.stats.size] size=+002 percent=+010 type=fileRobots",
		"INFO [db.stats.size] size=+003 percent=+015 type=fileHTML",
		"INFO [db.stats.size] size=+004 percent=+020 type=fileRSS",
		"INFO [db.stats.size] size=+005 percent=+025 type=fileSitemap",
		"INFO [db.stats.size] size=+006 percent=+030 type=fileFavicon",
	}, *records)
}
