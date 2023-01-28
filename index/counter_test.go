package index

import (
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCounterPage(t *testing.T) {
	count := CounterPage(0)
	count.Process(nil)
	count.Process(nil)
	assert.EqualValues(t, 2, count)
}

func TestCounterVocab(t *testing.T) {
	root, err := htmlnode.Parse([]byte("<p>  Hello WORD!\nHELLO yolo</p>"))
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
		WordFrequency{"hello", 4},
		WordFrequency{"word", 2},
		WordFrequency{"yolo", 2},
	}, counter.Frequency())
}
