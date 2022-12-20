package crawldatabase

import (
	"github.com/stretchr/testify/assert"
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
