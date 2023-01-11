package index

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/index/database"
	"strings"
)

type VocabAdvanced map[crawldatabase.Key]VocabAdvancedSlice
type VocabAdvancedSlice []VocabAdvancedItem
type VocabAdvancedItem struct {
	Page crawldatabase.Key
	Coef int32
}

func (advanced VocabAdvanced) Process(page *crawler.Page) {
	counter := make(map[string]int32)

	page.Html.Body.Visit(func(node htmlnode.Node) {
		text := node.Text
		text = strings.TrimSpace(text)
		text = strings.ToLower(text)
		if text == "" {
			return
		}
		for _, word := range strings.FieldsFunc(text, splitWords) {
			counter[word]++
		}
	})

	key := crawldatabase.NewKeyURL(&page.URL)
	for word, coef := range counter {
		wordKey := crawldatabase.NewKeyString(word)
		advanced[wordKey] = append(advanced[wordKey], VocabAdvancedItem{key, coef})
	}
}

/* STORE & LOAD */

func (slice VocabAdvancedSlice) MarshalBinary() ([]byte, error) {
	const itemLen = crawldatabase.KeyLen + 4
	data := make([]byte, len(slice)*itemLen)
	for i, value := range slice {
		copy(data[i*itemLen:], value.Page[:])
		data[i*itemLen+crawldatabase.KeyLen+0] = byte(value.Coef >> 24)
		data[i*itemLen+crawldatabase.KeyLen+1] = byte(value.Coef >> 16)
		data[i*itemLen+crawldatabase.KeyLen+2] = byte(value.Coef >> 8)
		data[i*itemLen+crawldatabase.KeyLen+3] = byte(value.Coef >> 0)
	}
	return data, nil
}

func LoadVocabAdvanced(file string) (VocabAdvanced, error) {
	return indexdatabase.Load(file, unmarshalVocabAdvancedSlice)
}

func unmarshalVocabAdvancedSlice(data []byte) (VocabAdvancedSlice, error) {
	const itemLen = crawldatabase.KeyLen + 4
	if len(data)%itemLen != 0 {
		return nil, fmt.Errorf("Expected len data multiple of %d, get: %d", itemLen, len(data))
	}

	slice := make(VocabAdvancedSlice, len(data)/itemLen)

	for i := range slice {
		key := crawldatabase.Key{}
		copy(key[:], data[i*itemLen:])

		p := i*itemLen + crawldatabase.KeyLen
		coef := 0 |
			int32(data[p+0])<<24 |
			int32(data[p+1])<<16 |
			int32(data[p+2])<<8 |
			int32(data[p+3])<<0

		slice[i] = VocabAdvancedItem{key, coef}
	}

	return slice, nil
}
