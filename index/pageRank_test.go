package index

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
)

var (
	pageA = keys.NewString("page:a")
	pageB = keys.NewString("page:b")
	pageC = keys.NewString("page:c")
	pageD = keys.NewString("page:d")
	pageX = keys.NewString("page:x")
	pageY = keys.NewString("page:y")
	pageR = keys.NewString("page:r")
)

func keyStringer(key keys.Key) string {
	switch key {
	case pageA:
		return "<pageA>"
	case pageB:
		return "<pageB>"
	case pageC:
		return "<pageC>"
	case pageD:
		return "<pageD>"
	case pageX:
		return "<pageX>"
	case pageY:
		return "<pageY>"
	case pageR:
		return "<pageR>"
	default:
		return fmt.Sprintf("<Unknown key:%x>", key[:])
	}
}

func TestLinksAddURLs(t *testing.T) {
	links := NewLinks(map[keys.Key]keys.Key{pageR: pageC})
	links.addURLs(common.ParseURL("page:a"), map[keys.Key]*url.URL{
		pageB: nil,
		pageR: nil,
	})
	assert.Equal(t, Links{
		redirection: map[keys.Key]keys.Key{pageR: pageC},
		globalLinks: map[keys.Key][]keys.Key{
			pageA: {pageB, pageC},
		},
	}, links)
}

func TestPageRank(t *testing.T) {
	repeat, scores := pageRank(map[keys.Key][]keys.Key{
		pageA: {pageC},
		pageB: {pageA, pageX},
		pageC: {pageA, pageY},
		pageD: {pageB, pageC, pageD, pageY},
	}, 1, 0.0)
	assert.Equal(t, 1, repeat)

	expected := map[keys.Key]float32{
		pageA: 2.0,
		pageC: 1.5,
		pageB: 0.5,
		pageD: 0.0,
	}

	if !reflect.DeepEqual(expected, scores) {
		t.Fail()
		printScores := func(name string, scores map[keys.Key]float32) {
			t.Log(name)
			for key, rank := range scores {
				t.Log(keyStringer(key), rank)
			}
			t.Log()
		}
		printScores("received:", scores)
		printScores("expected:", expected)
	}
}

func TestPageRankFilter(t *testing.T) {
	allLinks := map[keys.Key][]keys.Key{
		pageA: {pageC},
		pageB: {pageA, pageX},
		pageC: {pageA, pageY},
		pageD: {pageB, pageD, pageY},
	}
	pageRankFilter(allLinks)

	expected := map[keys.Key][]keys.Key{
		pageA: {pageC},
		pageB: {pageA},
		pageC: {pageA},
		pageD: {pageB},
	}
	if !reflect.DeepEqual(expected, allLinks) {
		print := func(m map[keys.Key][]keys.Key) {
			t.Logf("map len: %d", len(m))
			for key, links := range m {
				t.Log("\tkey:", keyStringer(key))
				for _, link := range links {
					t.Log("\t\t-", keyStringer(link))
				}
			}
		}

		t.Error("NOT EQUAL!")
		t.Log("")
		t.Log("expected:")
		print(expected)
		t.Log("")
		t.Log("received:")
		print(allLinks)
	}
}

func TestPageRankMultiplication(t *testing.T) {
	pages := [][]int{
		0: {2},
		1: {0},
		2: {0},
		3: {1, 2},
	}
	repeat, rank := pageRankMultiplication(pages, 1, 0.0)
	assert.Equal(t, []float32{2.0, 0.5, 1.5, 0.0}, rank)
	assert.Equal(t, 1, repeat)

	repeat, rank = pageRankMultiplication(pages, 2, 0.0)
	assert.Equal(t, []float32{2.0, 0, 2.0, 0}, rank)
	assert.Equal(t, 2, repeat)

	for i := 3; i < 100; i++ {
		repeat, rank = pageRankMultiplication(pages, i, 0.0)
		assert.Equal(t, []float32{2.0, 0, 2.0, 0}, rank)
		assert.Equal(t, 3, repeat)
	}
}

func TestRWPageRank(t *testing.T) {
	defer os.Remove("_pagerank.db")

	assert.NoError(t, StorePageRank("_pagerank.db", map[keys.Key]float32{
		pageA: 0.1,
		pageB: 1596.163541,
	}))

	scores, err := LoadPageRank("_pagerank.db")
	assert.NoError(t, err)
	assert.Equal(t, map[keys.Key]float32{
		pageA: 0.1,
		pageB: 1596.163541,
	}, scores)
}
