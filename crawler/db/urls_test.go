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
	assert.Equal(t, []string{
		"https://google.com/",
		"https://www.google.com/",
		"https://www.google.com/maps",
		"https://www.google.com/search",
		"https://www.google.com/travel",
	}, urls2str(urls))

	// Merge
	urls, err = db.Merge(inputURLS)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(urls), urls)
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://google.com/")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	// Store keys
	assert.NoError(t, db.store(NewStringKey("https://www.google.com/maps"), 535, true))
	assert.NoError(t, db.store(NewStringKey("https://www.google.com/search"), 659, false))
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://google.com/")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/")])
	assert.Equal(t, int64(-535), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(659), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	// Reopen
	assert.NoError(t, db.Close(), "close db")
	db, urls, err = OpenURLsDB(".", 30)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"https://google.com/",
		"https://www.google.com/",
		"https://www.google.com/travel",
	}, urls2str(urls))

	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://google.com/")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/")])
	assert.Equal(t, int64(-535), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(659), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	urls, err = db.Merge(inputURLS)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(urls), urls)

	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://google.com/")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/")])
	assert.Equal(t, int64(-535), db.keysMap[NewStringKey("https://www.google.com/maps")])
	assert.Equal(t, int64(659), db.keysMap[NewStringKey("https://www.google.com/search")])
	assert.Equal(t, int64(1), db.keysMap[NewStringKey("https://www.google.com/travel")])

	forOne := 0
	err = db.ForDone(func(key Key, _, _ int) error {
		forOne++
		assert.Equal(t, 1, forOne, "f executed only once.")
		assert.Equal(t, NewStringKey("https://www.google.com/search"), key)
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, forOne, "f executed only once.")

	assert.NoError(t, db.Close(), "close db")
}

func TestGetParentDomain(t *testing.T) {
	assert.Equal(t, []string{
		"https://isty.uvsq.fr/",
		"https://uvsq.fr/",
		"https://www.isty.uvsq.fr/",
		"https://www.isty.uvsq.fr:8000/",
		"https://www.isty.uvsq.fr:8000/dir/",
		"https://www.isty.uvsq.fr:8000/dir/subdir/",
		"https://www.isty.uvsq.fr:8000/dir/subdir/file.txt",
	}, urls2str(getParentURL(mustParse("https://www.isty.uvsq.fr:8000/dir/subdir/file.txt"))))

	assert.Equal(t, []string{
		"https://isty.uvsq.fr/",
		"https://uvsq.fr/",
		"https://www.isty.uvsq.fr/",
		"https://www.isty.uvsq.fr:8000/",
		"https://www.isty.uvsq.fr:8000/dir/",
		"https://www.isty.uvsq.fr:8000/dir/",
	}, urls2str(getParentURL(mustParse("https://www.isty.uvsq.fr:8000/dir/"))))

	assert.Equal(t, []string{
		"https://isty.uvsq.fr/",
		"https://uvsq.fr/",
		"https://www.isty.uvsq.fr/",
		"https://www.isty.uvsq.fr:8000/",
	}, urls2str(getParentURL(mustParse("https://www.isty.uvsq.fr:8000/"))))

	assert.Equal(t, []string{
		"https://isty.uvsq.fr/",
		"https://uvsq.fr/",
		"https://www.isty.uvsq.fr/",
		"https://www.isty.uvsq.fr/dir/",
		"https://www.isty.uvsq.fr/dir/file.txt",
	}, urls2str(getParentURL(mustParse("https://www.isty.uvsq.fr/dir/file.txt"))))

	assert.Equal(t, []string{
		"https://isty.uvsq.fr/",
		"https://uvsq.fr/",
		"https://www.isty.uvsq.fr/",
	}, urls2str(getParentURL(mustParse("https://www.isty.uvsq.fr/"))))

	assert.Equal(t, []string{
		"https://uvsq.fr/",
		"https://uvsq.fr/dir/",
		"https://uvsq.fr/dir/file.txt",
	}, urls2str(getParentURL(mustParse("https://uvsq.fr/dir/file.txt"))))

	assert.Equal(t, []string{
		"https://fr/",
		"https://fr/dir/",
		"https://fr/dir/file.txt",
	}, urls2str(getParentURL(mustParse("https://fr/dir/file.txt"))))
}

func urls2str(urls []*url.URL) []string {
	out := make([]string, len(urls))
	for i, u := range urls {
		if u == nil {
			out[i] = "<nil URL>"
		} else {
			out[i] = u.String()
		}
	}
	sort.Strings(out)
	return out
}
