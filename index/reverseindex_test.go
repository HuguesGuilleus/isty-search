package index

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestReverseIndexProcess(t *testing.T) {
	ri := make(ReverseIndex, 1)
	ri.Process(&crawler.Page{
		URL: *common.ParseURL("https://example.com/a"),
		Html: &htmlnode.Root{Body: htmlnode.Node{
			Text: "WORDA",
		}},
	})
	ri.Process(&crawler.Page{
		URL: *common.ParseURL("https://example.com/b"),
		Html: &htmlnode.Root{Body: htmlnode.Node{
			Text: " WordB\nWORDA wordA",
		}},
	})
	ri.Sort()
	assert.Equal(t, ReverseIndex{
		keys.NewString("worda"): []KeyFloat32{
			{keys.NewString("https://example.com/a"), 1.0},
			{keys.NewString("https://example.com/b"), 2.0},
		},
		keys.NewString("wordb"): []KeyFloat32{
			{keys.NewString("https://example.com/b"), 1.0},
		},
	}, ri)
}

func TestRverseIndexRW(t *testing.T) {
	defer os.Remove("_reverseindex.db")

	assert.NoError(t, ReverseIndex{
		keys.NewString("worda"): []KeyFloat32{
			{keys.NewString("https://example.com/a"), 1.0},
			{keys.NewString("https://example.com/b"), 2.0},
		},
		keys.NewString("wordb"): []KeyFloat32{
			{keys.NewString("https://example.com/b"), 1.0},
		},
	}.Store("_reverseindex.db"))

	ri, err := LoadReverseIndex("_reverseindex.db")
	assert.NoError(t, err)
	assert.Equal(t, ReverseIndex{
		keys.NewString("worda"): []KeyFloat32{
			{keys.NewString("https://example.com/a"), 1.0},
			{keys.NewString("https://example.com/b"), 2.0},
		},
		keys.NewString("wordb"): []KeyFloat32{
			{keys.NewString("https://example.com/b"), 1.0},
		},
	}, ri)
}
