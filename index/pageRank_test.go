package index

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	pageA = keys.NewString("page:a")
	pageB = keys.NewString("page:b")
	pageC = keys.NewString("page:c")
	pageD = keys.NewString("page:d")
	pageX = keys.NewString("page:x")
	pageY = keys.NewString("page:y")
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
	default:
		return fmt.Sprintf("<Unknown key:%x>", key[:])
	}
}

func TestPageRank(t *testing.T) {
	pr := PageRank{
		links: map[keys.Key][]keys.Key{
			pageA: {pageC},
			pageB: {pageA, pageX, pageA},
			pageC: {pageA, pageY},
			pageD: {pageB, pageC, pageD, pageY},
		},
	}
	scores := pr.Score(1)
	expected := []Score{
		Score{Key: pageA, Rank: 2.0},
		Score{Key: pageC, Rank: 1.5},
		Score{Key: pageB, Rank: 0.5},
		Score{Key: pageD, Rank: 0.0},
	}

	if !reflect.DeepEqual(expected, scores) {
		t.Fail()
		printScores := func(name string, scores []Score) {
			t.Log(name)
			for _, score := range scores {
				t.Log(keyStringer(score.Key), score.Rank)
			}
			t.Log()
		}
		printScores("received:", scores)
		printScores("expected:", expected)
	}
}

func TestPageRankFilter(t *testing.T) {
	pr := PageRank{
		links: map[keys.Key][]keys.Key{
			pageA: {pageC},
			pageB: {pageA, pageX, pageA},
			pageC: {pageA, pageY},
			pageD: {pageB, pageD, pageY},
		},
	}
	pr.filterKey()

	expected := map[keys.Key][]keys.Key{
		pageA: {pageC},
		pageB: {pageA},
		pageC: {pageA},
		pageD: {pageB},
	}
	if !reflect.DeepEqual(expected, pr.links) {
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
		print(pr.links)
	}
}

func TestPageRankMultiplication(t *testing.T) {
	pages := [][]int{
		0: []int{2},
		1: []int{0},
		2: []int{0},
		3: []int{1, 2},
	}
	assert.Equal(t, []float32{2.0, 0.5, 1.5, 0.0}, pageRankMultiplication(pages, 1))
	for i := 2; i < 10; i++ {
		assert.Equal(t, []float32{2.0, 0, 2.0, 0}, pageRankMultiplication(pages, i))
	}
}
