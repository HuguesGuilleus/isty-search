package crawler

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	database "github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/stretchr/testify/assert"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
)

// TODO: Move to a global package
func mustParse(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func mustParseSlice(args ...string) []*url.URL {
	urls := make([]*url.URL, len(args))
	for i, s := range args {
		u, err := url.Parse(s)
		if err != nil {
			panic(fmt.Sprintf("Wrong syntax for %q on index %d: %v", s, i, err))
		}
		urls[i] = u
	}
	return urls
}

func TestCrawl(t *testing.T) {
	defer os.RemoveAll("_testdb")
	db, urls, err := OpenDB("_testdb")
	assert.NoError(t, err)
	assert.Empty(t, urls)
	defer db.Close()

	err = Crawl(context.Background(), Config{
		DB:    db,
		Input: []*url.URL{mustParse("https://example.org/")},
		FilterURL: []func(*url.URL) string{func(u *url.URL) string {
			if u.Host != "example.org" {
				return "wrong host"
			}
			return ""
		}},
		FilterPage: []func(*htmlnode.Root) string{func(page *htmlnode.Root) string {
			if page.Meta.Langage != "en" {
				return "wring langage"
			}
			return ""
		}},
		MaxLength:    15_000_000,
		MaxGo:        1,
		RoundTripper: datatestRoundTripper{},
	})
	assert.NoError(t, err)

	// Test URLs was fetched
	urls, err = db.URLsDB.Merge(mustParseSlice(
		"https://example.org/",
		"https://example.org/dir/",
		"https://example.org/dir/subdir/",
		"https://example.org/es.html",
		"https://example.org/robotBlocked.html",
		"https://google.com/",
	))
	assert.NoError(t, err)
	assert.Empty(t, urls)

	// Test page
	paths := []string{
		"/",
		"/dir/",
		"/dir/subdir/",
	}
	for _, path := range paths {
		data, err := fs.ReadFile(datatest, "datatest"+path+"index.html")
		assert.NoError(t, err)
		root, err := htmlnode.Parse(data)
		assert.NoError(t, err)

		page, err := db.KeyValueDB.Get(database.NewStringKey("https://example.org" + path))
		assert.NoError(t, err, "https://example.org"+path)
		assert.Equal(t, root.Head.PrintLines(), page.Html.Head.PrintLines())
		assert.Equal(t, root.Body.PrintLines(), page.Html.Body.PrintLines())
	}
}

//go:embed datatest
var datatest embed.FS

type datatestRoundTripper struct{}

func (_ datatestRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.Method != http.MethodGet {
		panic("Method not allowed")
	}

	if request.Host != "example.org" {
		panic("no example.com host")
	}

	path := request.URL.Path
	if strings.HasSuffix(path, "/") {
		path += "index.html"
	}
	data, err := fs.ReadFile(datatest, "datatest"+path)
	if err != nil {
		panic("Not found:" + path)
	}

	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,

		Proto:      request.Proto,
		ProtoMajor: request.ProtoMajor,
		ProtoMinor: request.ProtoMinor,

		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: int64(len(data)),
		Request:       request,
	}, nil
}
