package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"testing"
	"time"
)

func TestDBFile(t *testing.T) {
	// Web use http.Cookie because it's a struct.
	expectedValue := http.Cookie{
		Name:    "yolo",
		Value:   "swag",
		Expires: time.Date(2022, time.October, 5, 20, 8, 50, 0, time.UTC),
	}
	key := keys.NewString("key")

	defer os.RemoveAll("__db")

	records, handler := sloghandlers.NewHandlerRecords(slog.InfoLevel)

	urls, db, err := OpenWithKnow[http.Cookie](slog.New(handler), "__db", false)
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
	v, storedTime, err := db.GetValue(key)
	assert.Nil(t, v)
	assert.True(t, storedTime.IsZero())
	assert.Error(t, err)

	// Set
	assert.NoError(t, db.SetValue(key, &http.Cookie{}, TypeFileHTML))
	assert.NoError(t, db.SetValue(key, &expectedValue, TypeFileHTML))
	assert.NoError(t, db.SetValue(keys.NewString("qsdgfd"), &http.Cookie{}, TypeFileHTML))
	assert.NoError(t, db.SetValue(keys.NewString("gejkhk"), &http.Cookie{}, TypeFileHTML))

	// Get
	v, storedTime, err = db.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, &expectedValue, v)
	assert.False(t, storedTime.IsZero())

	// Set simple
	ks := keys.NewString("simple")
	assert.Error(t, db.SetSimple(ks, TypeFileRSS))
	assert.NoError(t, db.SetSimple(ks, TypeErrorParsing))

	meta := db.mapMeta[ks]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeErrorParsing}, meta)

	// Remove key
	kd := keys.NewString("deleted")
	assert.NoError(t, db.SetSimple(kd, TypeErrorParsing))
	assert.NotZero(t, db.mapMeta[kd])
	assert.NoError(t, db.SetSimple(kd, TypeNothing))
	assert.Zero(t, db.mapMeta[kd])

	// Redirection
	ko := keys.NewString("origin")
	kt := keys.NewString("target")
	assert.NoError(t, db.SetRedirect(ko, kt))
	meta = db.mapMeta[ko]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeRedirect, Hash: kt}, meta)

	// Reopen
	assert.NoError(t, db.Close())
	urls, db, err = OpenWithKnow[http.Cookie](slog.New(handler), "__db", false)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"https://google.com",
		"https://www.google.com",
	}, common.URL2String(urls))

	urlsMap, _ = getURLS()
	assert.NoError(t, db.AddURL(urlsMap))
	assert.Empty(t, urlsMap)

	// Get
	v, storedTime, err = db.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, &expectedValue, v)
	assert.False(t, storedTime.IsZero())

	// Check simple
	assert.Zero(t, db.mapMeta[kd])
	meta = db.mapMeta[ks]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeErrorParsing}, meta)

	// Check redirect
	meta = db.mapMeta[ko]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeRedirect, Hash: kt}, meta)

	// Redirections
	koo := keys.NewString("origin-origin")
	assert.NoError(t, db.SetValue(kt, &http.Cookie{}, TypeFile))
	assert.NoError(t, db.SetRedirect(koo, ko))
	assert.NoError(t, db.SetRedirect(keys.NewString("2null"), keys.NewString("null")))
	// the chain koo -> ko -> kt
	assert.Equal(t, map[keys.Key]keys.Key{
		koo: kt,
		ko:  kt,
	}, db.Redirections())

	// Reopen without urls.
	assert.NoError(t, db.Close())
	urls, db, err = Open[http.Cookie](slog.New(handler), "__db", false)
	assert.Nil(t, urls)
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	// No log
	assert.Equal(t, []string{
		"INFO [db.open] base=__db",
		"INFO [db.open] base=__db",
		"INFO [db.open] base=__db",
	}, *records)
}

func TestForHTML(t *testing.T) {
	cookie1 := &http.Cookie{Name: "1", MaxAge: 1}
	cookie2 := &http.Cookie{Name: "2", MaxAge: 2}
	cookie3 := &http.Cookie{Name: "3", MaxAge: 3}

	k1 := keys.NewString("k1")
	k2 := keys.NewString("k2")
	k3 := keys.NewString("k3")

	defer os.RemoveAll("__db")
	records, handler := sloghandlers.NewHandlerRecords(slog.InfoLevel)
	urls, db, err := Open[http.Cookie](slog.New(handler), "__db", false)
	assert.Nil(t, urls)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, db.Close()) }()

	// Add values
	assert.NoError(t, db.SetSimple(keys.NewString("err"), TypeErrorNetwork))
	assert.NoError(t, db.SetValue(k1, cookie1, TypeFileHTML))
	assert.NoError(t, db.SetValue(k2, cookie2, TypeFileHTML))
	assert.NoError(t, db.SetValue(k3, cookie3, TypeFileHTML))

	// Check ForHTML caller
	readed := make(map[keys.Key]*http.Cookie, 3)
	globalInc := 0
	err = db.ForHTML(func(key keys.Key, c *http.Cookie, i, total int) {
		assert.Equal(t, 3, total)
		assert.Nil(t, readed[key])
		readed[key] = c
		assert.Equal(t, i, c.MaxAge-1)
		globalInc++
	})
	assert.NoError(t, err)
	assert.Equal(t, map[keys.Key]*http.Cookie{
		k1: cookie1,
		k2: cookie2,
		k3: cookie3,
	}, readed)

	// Log records
	sort.Strings(*records)
	assert.Equal(t, []string{
		"INFO [%] %i=+000 %len=+003",
		"INFO [%] %i=+001 %len=+003",
		"INFO [%] %i=+002 %len=+003",
		"INFO [%end]",
		"INFO [db.open] base=__db",
	}, *records)
}

func TestDBMemory(t *testing.T) {
	// Web use http.Cookie because it's a struct.
	expectedValue := http.Cookie{
		Name:    "yolo",
		Value:   "swag",
		Expires: time.Date(2022, time.October, 5, 20, 8, 50, 0, time.UTC),
	}
	key := keys.NewString("key")

	records, handler := sloghandlers.NewHandlerRecords(slog.InfoLevel)
	urls, db, err := OpenMemory[http.Cookie](slog.New(handler), "", false)
	assert.Nil(t, urls)
	assert.NoError(t, err)

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
	v, storedTime, err := db.GetValue(key)
	assert.Nil(t, v)
	assert.True(t, storedTime.IsZero())
	assert.Error(t, err)

	// Set
	assert.NoError(t, db.SetValue(key, &http.Cookie{}, TypeFileHTML))
	assert.NoError(t, db.SetValue(key, &expectedValue, TypeFileHTML))
	assert.NoError(t, db.SetValue(keys.NewString("qsdgfd"), &http.Cookie{}, TypeFileHTML))
	assert.NoError(t, db.SetValue(keys.NewString("gejkhk"), &http.Cookie{}, TypeFileHTML))

	// Get
	v, storedTime, err = db.GetValue(key)
	assert.NoError(t, err)
	assert.Equal(t, &expectedValue, v)
	assert.False(t, storedTime.IsZero())

	// Set simple
	ks := keys.NewString("simple")
	assert.Error(t, db.SetSimple(ks, TypeFileRSS))
	assert.NoError(t, db.SetSimple(ks, TypeErrorParsing))

	meta := db.mapMeta[ks]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeErrorParsing}, meta)

	// Remove key
	kd := keys.NewString("deleted")
	assert.NoError(t, db.SetSimple(kd, TypeErrorParsing))
	assert.NotZero(t, db.mapMeta[kd])
	assert.NoError(t, db.SetSimple(kd, TypeNothing))
	assert.Zero(t, db.mapMeta[kd])

	// Redirection
	ko := keys.NewString("origin")
	kt := keys.NewString("target")
	assert.NoError(t, db.SetRedirect(ko, kt))
	meta = db.mapMeta[ko]
	assert.NotZero(t, meta.Time)
	meta.Time = 0
	assert.Equal(t, metavalue{Type: TypeRedirect, Hash: kt}, meta)

	assert.Nil(t, *records)
}

func getURLS() (map[keys.Key]*url.URL, int) {
	originURLS := common.ParseURLs(
		"https://google.com",
		"https://www.google.com",
	)
	urlsMap := make(map[keys.Key]*url.URL)
	for _, u := range originURLS {
		urlsMap[keys.NewURL(u)] = u
	}
	return urlsMap, len(urlsMap)
}
