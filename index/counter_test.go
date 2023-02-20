package index

import (
	"testing"

	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/stretchr/testify/assert"
)

func TestCounterVocab(t *testing.T) {
	root, err := htmlnode.Parse([]byte("<p> a \t   bb  0 Hello WORD!\nHELLO yolo fffffffffffff</p>"))
	assert.NoError(t, err)

	counter := make(CounterVocab)
	counter.Process(&crawler.Page{Html: root})
	counter.Process(&crawler.Page{Html: root})

	assert.Equal(t, CounterVocab{
		"hello": 4,
		"word":  2,
		"yolo":  2,
	}, counter)
	assert.Equal(t, 8, counter.Sum())
	assert.Equal(t, []WordFrequency{
		{"hello", 4},
		{"word", 2},
		{"yolo", 2},
	}, counter.Frequency())
}
