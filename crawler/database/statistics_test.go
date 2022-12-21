package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestGetStatistics(t *testing.T) {
	assert.Equal(t, Statistics{}, getStatistics(map[Key]metavalue{}))
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
	}, getStatistics(map[Key]metavalue{
		NewKeyString("key0"): metavalue{Type: TypeKnow},
		NewKeyString("key1"): metavalue{Type: TypeRedirect},

		NewKeyString("key2"): metavalue{Type: TypeFileRobots, Length: 2},
		NewKeyString("key3"): metavalue{Type: TypeFileHTML, Length: 3},
		NewKeyString("key4"): metavalue{Type: TypeFileRSS, Length: 4},
		NewKeyString("key5"): metavalue{Type: TypeFileSitemap, Length: 5},
		NewKeyString("key6"): metavalue{Type: TypeFileFavicon, Length: 6},

		NewKeyString("key7"):  metavalue{Type: TypeErrorNetwork},
		NewKeyString("key8"):  metavalue{Type: TypeErrorParsing},
		NewKeyString("key9"):  metavalue{Type: TypeErrorFilterURL},
		NewKeyString("key10"): metavalue{Type: TypeErrorFilterPage},
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
		"INFO [db.stats.total] count.all=11 count.file=5 count.error=4 size=20",
	}, *records)

	records, handler = sloghandlers.NewHandlerRecords(slog.DebugLevel)
	stats.LogAll(slog.New(handler))
	assert.Equal(t, []string{
		"INFO [db.stats.total] count.all=11 count.file=5 count.error=4 size=20",
		"INFO [db.stats.count] count=1 percent=9 type=know",
		"INFO [db.stats.count] count=1 percent=9 type=redirect",
		"INFO [db.stats.count] count=1 percent=9 type=fileRobots",
		"INFO [db.stats.count] count=1 percent=9 type=fileHTML",
		"INFO [db.stats.count] count=1 percent=9 type=fileRSS",
		"INFO [db.stats.count] count=1 percent=9 type=fileSitemap",
		"INFO [db.stats.count] count=1 percent=9 type=fileFavicon",
		"INFO [db.stats.count] count=1 percent=9 type=errorNetwork",
		"INFO [db.stats.count] count=1 percent=9 type=errorParsing",
		"INFO [db.stats.count] count=1 percent=9 type=errorFilterURL",
		"INFO [db.stats.count] count=1 percent=9 type=errorFilterPage",
		"INFO [db.stats.size] total=20",
		"INFO [db.stats.size] size=2 percent=10 type=fileRobots",
		"INFO [db.stats.size] size=3 percent=15 type=fileHTML",
		"INFO [db.stats.size] size=4 percent=20 type=fileRSS",
		"INFO [db.stats.size] size=5 percent=25 type=fileSitemap",
		"INFO [db.stats.size] size=6 percent=30 type=fileFavicon",
	}, *records)
}
