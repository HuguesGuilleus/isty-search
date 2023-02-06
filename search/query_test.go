package search

import (
	"github.com/HuguesGuilleus/isty-search/index"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSearch(t *testing.T) {
	queries, pages := search("hello WORD", map[keys.Key][]index.KeyFloat32{
		keys.NewString("hello"): []index.KeyFloat32{
			index.KeyFloat32{keys.Key{1}, 0},
			index.KeyFloat32{keys.Key{3}, 0},
			index.KeyFloat32{keys.Key{4}, 0},
			index.KeyFloat32{keys.Key{5}, 0},
		},
		keys.NewString("word"): []index.KeyFloat32{
			index.KeyFloat32{keys.Key{1}, 0},
			index.KeyFloat32{keys.Key{2}, 0},
			index.KeyFloat32{keys.Key{3}, 0},
			index.KeyFloat32{keys.Key{5}, 0},
		},
	})
	assert.Equal(t, []Query{
		Query{"hello", keys.NewString("hello"), 4},
		Query{"word", keys.NewString("word"), 4},
	}, queries)
	assert.Equal(t, []index.KeyFloat32{
		index.KeyFloat32{keys.Key{1}, 0},
		index.KeyFloat32{keys.Key{3}, 0},
		index.KeyFloat32{keys.Key{5}, 0},
	}, pages)

	_, pages = search("hello WORD", map[keys.Key][]index.KeyFloat32{})
	assert.Nil(t, pages)

	queries, pages = search("", map[keys.Key][]index.KeyFloat32{})
	assert.Nil(t, queries)
	assert.Nil(t, pages)
}

func TestParse(t *testing.T) {
	assert.Equal(t, []Query{
		Query{"hello", keys.NewString("hello"), 0},
		Query{"word", keys.NewString("word"), 0},
	}, parse("HELLO WORD!"))
	assert.Nil(t, parse("aa"))
}

func TestMergeKeyFloat32(t *testing.T) {
	assert.Equal(t, []index.KeyFloat32{
		index.KeyFloat32{keys.Key{1}, 1.1},
		index.KeyFloat32{keys.Key{3}, 1.3},
		index.KeyFloat32{keys.Key{6}, 1.6},
	}, commonKeyFloat32s([]index.KeyFloat32{
		index.KeyFloat32{keys.Key{1}, 1.1},
		index.KeyFloat32{keys.Key{2}, 1.2},
		index.KeyFloat32{keys.Key{3}, 1.3},
		index.KeyFloat32{keys.Key{4}, 1.4},
		index.KeyFloat32{keys.Key{6}, 1.6},
		index.KeyFloat32{keys.Key{7}, 1.7},
	}, []index.KeyFloat32{
		index.KeyFloat32{keys.Key{0}, 1.0},
		index.KeyFloat32{keys.Key{1}, 1.1},
		index.KeyFloat32{keys.Key{3}, 1.3},
		index.KeyFloat32{keys.Key{5}, 1.5},
		index.KeyFloat32{keys.Key{6}, 1.6},
	}))
}

func TestScore(t *testing.T) {
	pages := []index.KeyFloat32{
		index.KeyFloat32{keys.Key{0}, 1.0},
		index.KeyFloat32{keys.Key{1}, 1.0},
		index.KeyFloat32{keys.Key{2}, 1.0},
	}
	score(pages, map[keys.Key]float32{
		keys.Key{0}: 0.0,
		keys.Key{1}: 0.1,
		keys.Key{2}: 0.2,
	})
	assert.Equal(t, []index.KeyFloat32{
		index.KeyFloat32{keys.Key{2}, 1.2},
		index.KeyFloat32{keys.Key{1}, 1.1},
		index.KeyFloat32{keys.Key{0}, 1.0},
	}, pages)
}
