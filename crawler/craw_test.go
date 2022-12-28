package crawler

import (
	"bytes"
	"context"
	"embed"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
)

func TestCrawl(t *testing.T) {
	records, logHandler := sloghandlers.NewHandlerRecords(slog.DebugLevel)
	logger := slog.New(logHandler)

	_, db, _ := crawldatabase.OpenMemory[Page](logger, "", false)
	err := Crawl(context.Background(), Config{
		DBopener: func(argLogger *slog.Logger, base string, logStatistics bool) ([]*url.URL, crawldatabase.Database[Page], error) {
			assert.Equal(t, logger, argLogger)
			assert.Equal(t, "$db-base$", base)
			assert.True(t, logStatistics)
			return nil, db, nil
		},
		DBbase: "$db-base$",

		Input: []*url.URL{common.ParseURL("https://example.org/")},
		FilterURL: []func(*url.URL) bool{
			func(u *url.URL) bool { return u.Host != "example.org" },
		},
		FilterPage: []func(*htmlnode.Root) bool{
			func(page *htmlnode.Root) bool { return page.Meta.Langage != "en" },
		},
		MaxLength:    15_000_000,
		MaxGo:        1,
		RoundTripper: datatestRoundTripper{},
		Logger:       logger,
	})
	assert.NoError(t, err)

	// Test log records
	sort.Strings(*records)
	assert.Equal(t, []string{
		`INFO [fetch.ok] status=200 url=https://example.org/`,
		`INFO [fetch.ok] status=200 url=https://example.org/dir/`,
		`INFO [fetch.ok] status=200 url=https://example.org/dir/subdir/`,
		`INFO [fetch.ok] status=200 url=https://example.org/es.html`,
		`INFO [fetch.ok] status=200 url=https://example.org/robots.txt`,
	}, *records)

	// Test page
	paths := []string{
		"/",
		"/dir/",
		"/dir/subdir/",
	}
	for _, path := range paths {
		data, err := fs.ReadFile(testdata, "testdata"+path+"index.html")
		assert.NoError(t, err)
		root, err := htmlnode.Parse(data)
		assert.NoError(t, err)

		page, _, err := db.GetValue(crawldatabase.NewKeyString("https://example.org" + path))
		assert.NoError(t, err, "https://example.org"+path)
		assert.Equal(t, root.Head.PrintLines(), page.Html.Head.PrintLines())
		assert.Equal(t, root.Body.PrintLines(), page.Html.Body.PrintLines())
	}

	// Test the process
	foundURL := []string{}
	assert.NoError(t, Process(db,
		slog.New(sloghandlers.NewNullHandler()),
		ProcessFunc(func(page *Page) {
			foundURL = append(foundURL, page.URL.String())
		}),
	))
	sort.Strings(foundURL)
	assert.Equal(t, []string{
		"https://example.org/",
		"https://example.org/dir/",
		"https://example.org/dir/subdir/",
	}, foundURL)
}

//go:embed testdata
var testdata embed.FS

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
	data, err := fs.ReadFile(testdata, "testdata"+path)
	if err != nil {
		panic("Not found: " + path)
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
