package crawler

import (
	_ "embed"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

//go:embed page_urls.html
var exampleURLHtml []byte

func TestGetURL(t *testing.T) {
	root, err := htmlnode.Parse(exampleURLHtml)
	assert.NoError(t, err)

	page := Page{
		URL:  *common.ParseURL("https://example.com/"),
		Html: root,
	}

	urls := make([]string, 0)
	for _, u := range page.GetURLs() {
		urls = append(urls, u.String())
	}
	sort.Strings(urls)
	assert.Equal(t, []string{
		"https://github.com/",
		"https://w1.w2.yolo.net/",
		"https://w1.w2.yolo.net:8000/",
		"https://w1.w2.yolo.net:8000/dir/",
		"https://w1.w2.yolo.net:8000/dir/subdir/",
		"https://w1.w2.yolo.net:8000/dir/subdir/swag",
		"https://w1.w2.yolo.net:8000/dir/subdir/swag?a=1&b=2",
		"https://w2.yolo.net/",
		"https://www.github.com/",
		"https://www.yolo.net/",
		"https://yolo.net/",
	}, urls)
}
