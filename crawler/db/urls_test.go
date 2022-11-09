package db

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"sort"
	"testing"
)

func TestURLsDB(t *testing.T) {
	defer os.Remove("urls.key")
	defer os.Remove("urls.txt")

	inputURLS := []*url.URL{
		mustParse("https://www.google.com/favicon.ico"),
		mustParse("https://www.google.com/maps"),
		mustParse("https://www.google.com/robots.txt"),
		mustParse("https://www.google.com/search"),
		mustParse("https://www.google.com/travel"),
		mustParse("https://www.google.com/verylongverylongverylongverylongverylong"),
	}

	db, urls, err := OpenURLsDB(".", 30)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(urls), urls)

	assert.Equal(t, int64(0), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(0), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(0), db.keysMap[NewStringKey("https://www.google.com/travel")])

	urls, err = db.Merge(inputURLS)
	assert.NoError(t, err)
	sort.Slice(urls, func(i, j int) bool { return urls[i].String() < urls[j].String() })
	assert.Equal(t, []*url.URL{
		mustParse("https://www.google.com/maps"),
		mustParse("https://www.google.com/search"),
		mustParse("https://www.google.com/travel"),
	}, urls)

	// Merge
	urls, err = db.Merge(inputURLS)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(urls), urls)
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	// Store keys
	assert.NoError(t, db.store(NewStringKey("https://www.google.com/maps"), 535, true))
	assert.NoError(t, db.store(NewStringKey("https://www.google.com/search"), 659, false))
	assert.Equal(t, int64(-535), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(659), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	// Reopen
	assert.NoError(t, db.Close(), "close db")
	db, urls, err = OpenURLsDB(".", 30)
	assert.NoError(t, err)
	assert.Equal(t, []*url.URL{
		mustParse("https://www.google.com/travel"),
	}, urls)

	assert.Equal(t, int64(-535), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(659), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	urls, err = db.Merge(inputURLS)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(urls), urls)

	assert.Equal(t, int64(-535), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(659), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	assert.NoError(t, db.Close(), "close db")
}
