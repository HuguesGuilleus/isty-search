package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
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
	// Web use http.Cookie because it's a struct.
	expectedValue := http.Cookie{
		Name:    "yolo",
		Value:   "swag",
		Expires: time.Date(2022, time.October, 5, 20, 8, 50, 0, time.UTC),
	}
	key := NewKeyString("key")

	defer os.RemoveAll("__db")

	records, handler := sloghandlers.NewHandlerRecords(slog.WarnLevel)

	urls, db, err := OpenWithKnow[http.Cookie](slog.New(handler), "__db")
	assert.NoError(t, err)
	assert.Empty(t, urls)

	// First AddURL
	urlsMap, originLen := getURLS()
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Len(t, urlsMap, originLen)

	// Second AddURL
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Empty(t, urlsMap)

	// Wrong set
	assert.Error(t, db.SetValue(key, &http.Cookie{}, TypeErrorNetwork))
	assert.Error(t, db.SetValue(key, nil, TypeFileHTML))
	v, err := db.GetValue(key)
	assert.Nil(t, v)
	assert.Error(t, err)

	// Set
	assert.NoError(t, db.SetValue(key, &http.Cookie{}, TypeFileHTML))
	assert.NoError(t, db.SetValue(key, &expectedValue, TypeFileHTML))
	assert.NoError(t, db.SetValue(NewKeyString("qsdgfd"), &http.Cookie{}, TypeFileHTML))
	assert.NoError(t, db.SetValue(NewKeyString("gejkhk"), &http.Cookie{}, TypeFileHTML))

	// Get
	v, err = db.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, &expectedValue, v)

	// Reopen
	assert.NoError(t, db.Close())
	urls, db, err = OpenWithKnow[http.Cookie](slog.New(handler), "__db")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"https://google.com",
		"https://www.google.com",
	}, common.URL2String(urls))

	urlsMap, _ = getURLS()
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Empty(t, urlsMap)

	// Get
	v, err = db.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, &expectedValue, v)

	assert.NoError(t, db.Close())

	// No log
	assert.Empty(t, *records)
}
