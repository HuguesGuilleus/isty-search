package search

import (
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	assert.Equal(t, []Query{
		Query{"hello", keys.NewString("hello"), 0},
		Query{"word", keys.NewString("word"), 0},
	}, parse("HELLO WORD!"))
}
