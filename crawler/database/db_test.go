package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
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

	// Set simple
	ks := NewKeyString("simple")
	assert.Error(t, db.SetSimple(ks, TypeFileRSS))
	assert.NoError(t, db.SetSimple(ks, TypeErrorParsing))

	meta := db.(*database[http.Cookie]).mapMeta[ks]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeErrorParsing}, meta)

	// Remove key
	kd := NewKeyString("deleted")
	assert.NoError(t, db.SetSimple(kd, TypeErrorParsing))
	assert.NotZero(t, db.(*database[http.Cookie]).mapMeta[kd])
	assert.NoError(t, db.SetSimple(kd, TypeNothing))
	assert.Zero(t, db.(*database[http.Cookie]).mapMeta[kd])

	// Redirection
	ko := NewKeyString("origin")
	kt := NewKeyString("target")
	assert.NoError(t, db.SetRedirect(ko, kt))
	meta = db.(*database[http.Cookie]).mapMeta[ko]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeRedirect, Hash: kt}, meta)

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

	// Check simple
	assert.Zero(t, db.(*database[http.Cookie]).mapMeta[kd])
	meta = db.(*database[http.Cookie]).mapMeta[ks]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeErrorParsing}, meta)

	// Check redirect
	meta = db.(*database[http.Cookie]).mapMeta[ko]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeRedirect, Hash: kt}, meta)

	// Redirections
	koo := NewKeyString("origin-origin")
	assert.NoError(t, db.SetValue(kt, &http.Cookie{}, TypeFile))
	assert.NoError(t, db.SetRedirect(koo, ko))
	assert.NoError(t, db.SetRedirect(NewKeyString("2null"), NewKeyString("null")))
	// the chain koo -> ko -> kt
	assert.Equal(t, map[Key]Key{
		koo: kt,
		ko:  kt,
	}, db.Redirections())

	// Reopen without urls.
	assert.NoError(t, db.Close())
	urls, db, err = Open[http.Cookie](slog.New(handler), "__db")
	assert.Nil(t, urls)
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	// No log
	assert.Empty(t, *records)
}

func TestForHTML(t *testing.T) {
	cookie1 := &http.Cookie{Name: "1", MaxAge: 1}
	cookie2 := &http.Cookie{Name: "2", MaxAge: 2}
	cookie3 := &http.Cookie{Name: "3", MaxAge: 3}

	k1 := NewKeyString("k1")
	k2 := NewKeyString("k2")
	k3 := NewKeyString("k3")

	defer os.RemoveAll("__db")
	records, handler := sloghandlers.NewHandlerRecords(slog.InfoLevel)
	urls, db, err := Open[http.Cookie](slog.New(handler), "__db")
	assert.Nil(t, urls)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, db.Close()) }()

	// Add values
	assert.NoError(t, db.SetSimple(NewKeyString("err"), TypeErrorNetwork))
	assert.NoError(t, db.SetValue(k1, cookie1, TypeFileHTML))
	assert.NoError(t, db.SetValue(k2, cookie2, TypeFileHTML))
	assert.NoError(t, db.SetValue(k3, cookie3, TypeFileHTML))

	// Check ForHTML caller
	readed := make(map[Key]*http.Cookie, 3)
	globalInc := 0
	err = db.ForHTML(func(key Key, c *http.Cookie, i, total int) {
		assert.Equal(t, 3, total)
		assert.Nil(t, readed[key])
		readed[key] = c
		assert.Equal(t, globalInc, i)
		assert.Equal(t, globalInc, c.MaxAge-1)
		globalInc++
	})
	assert.NoError(t, err)
	assert.Equal(t, map[Key]*http.Cookie{
		k1: cookie1,
		k2: cookie2,
		k3: cookie3,
	}, readed)

	// Log records
	iterRorecords := make([]string, 0, len(*records))
	for _, r := range *records {
		if strings.HasPrefix(r, "INFO [db.stats.total]") {
			continue
		} else if strings.HasPrefix(r, "INFO [db.open]") {
			continue
		}
		iterRorecords = append(iterRorecords, r)
	}
	assert.Equal(t, []string{
		"INFO [%] %i=0 %len=3",
		"INFO [%] %i=1 %len=3",
		"INFO [%] %i=2 %len=3",
		"INFO [%end]",
	}, iterRorecords)
}