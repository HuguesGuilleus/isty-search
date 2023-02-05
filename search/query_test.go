package search

import (
	"github.com/HuguesGuilleus/isty-search/index"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
