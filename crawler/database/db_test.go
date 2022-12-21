package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"net/url"
	"os"
	"testing"
)

func TestDB(t *testing.T) {
	getURLS := func() (map[Key]*url.URL, int) {
		originURLS := common.ParseURLs(
			"https://google.com",
			"https://www.google.com",
		)
		urlsMap := make(map[Key]*url.URL)
		for _, u := range originURLS {
			urlsMap[NewKeyURL(u)] = u
		}
		return urlsMap, len(urlsMap)
	}

	defer os.RemoveAll("__db")

	records, handler := sloghandlers.NewHandlerRecords(slog.WarnLevel)

	urls, db, err := OpenWithKnow[any](slog.New(handler), "__db")
	assert.NoError(t, err)
	assert.Empty(t, urls)

	// First AddURL
	urlsMap, originLen := getURLS()
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Len(t, urlsMap, originLen)

	// Second AddURL
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Empty(t, urlsMap)

	assert.NoError(t, db.Close())

	// Reopen
	urls, db, err = OpenWithKnow[any](slog.New(handler), "__db")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"https://google.com",
		"https://www.google.com",
	}, common.URL2String(urls))

	urlsMap, _ = getURLS()
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Empty(t, urlsMap)

	assert.NoError(t, db.Close())

	// No log
	assert.Empty(t, *records)
}
