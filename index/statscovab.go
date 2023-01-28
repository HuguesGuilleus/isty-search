package index

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/index/database"
	"github.com/HuguesGuilleus/isty-search/keys"
)

type VocabAdvanced map[keys.Key]VocabAdvancedSlice
type VocabAdvancedSlice []VocabAdvancedItem
type VocabAdvancedItem struct {
	Page keys.Key
	Coef int32
}

func (advanced VocabAdvanced) Process(page *crawler.Page) {
	counter := make(map[string]int32)

	page.Html.Body.Visit(func(node htmlnode.Node) {
		for _, word := range getVocab(node.Text) {
			counter[word]++
		}
	})

	key := keys.NewURL(&page.URL)
	for word, coef := range counter {
		wordKey := keys.NewString(word)
		advanced[wordKey] = append(advanced[wordKey], VocabAdvancedItem{key, coef})
	}
}

/* STORE & LOAD */

func (slice VocabAdvancedSlice) MarshalBinary() ([]byte, error) {
	const itemLen = keys.Len + 4
	data := make([]byte, len(slice)*itemLen)
	for i, value := range slice {
		copy(data[i*itemLen:], value.Page[:])
		data[i*itemLen+keys.Len+0] = byte(value.Coef >> 24)
		data[i*itemLen+keys.Len+1] = byte(value.Coef >> 16)
		data[i*itemLen+keys.Len+2] = byte(value.Coef >> 8)
		data[i*itemLen+keys.Len+3] = byte(value.Coef >> 0)
	}
	return data, nil
}

func LoadVocabAdvanced(file string) (VocabAdvanced, error) {
	return indexdatabase.Load(file, unmarshalVocabAdvancedSlice)
}
func unmarshalVocabAdvancedSlice(data []byte) (VocabAdvancedSlice, error) {
	const itemLen = keys.Len + 4
	if len(data)%itemLen != 0 {
		return nil, fmt.Errorf("Expected len data multiple of %d, get: %d", itemLen, len(data))
	}

	slice := make(VocabAdvancedSlice, len(data)/itemLen)

	for i := range slice {
		key := keys.Key{}
		copy(key[:], data[i*itemLen:])

		p := i*itemLen + keys.Len
		coef := 0 |
			int32(data[p+0])<<24 |
			int32(data[p+1])<<16 |
			int32(data[p+2])<<8 |
			int32(data[p+3])<<0

		slice[i] = VocabAdvancedItem{key, coef}
	}

	return slice, nil
}
