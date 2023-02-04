package index

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/index/database"
	"github.com/HuguesGuilleus/isty-search/keys"
	"math"
	"sort"
)

type ReverseIndex map[keys.Key][]KeyFloat32

type KeyFloat32 struct {
	keys.Key
	F32 float32
}

func (index ReverseIndex) Process(page *crawler.Page) {
	counter := make(map[string]float32)

	page.Html.Body.Visit(func(node htmlnode.Node) {
		for _, word := range getVocab(node.Text) {
			counter[word]++
		}
	})

	key := keys.NewURL(&page.URL)
	for word, coef := range counter {
		wordKey := keys.NewString(word)
		index[wordKey] = append(index[wordKey], KeyFloat32{key, coef})
	}
}

// Sort map item by the order of the key.
func (advanced ReverseIndex) Sort() {
	for _, items := range advanced {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Key.Less(&items[j].Key)
		})
	}
}

func (advanced ReverseIndex) Store(file string) error {
	return indexdatabase.Store(file, advanced, func(slice []KeyFloat32) []byte {
		const itemLen = keys.Len + 4
		data := make([]byte, len(slice)*itemLen)
		for i, value := range slice {
			copy(data[i*itemLen:], value.Key[:])
			u := math.Float32bits(value.F32)
			data[i*itemLen+keys.Len+0] = byte(u >> 24)
			data[i*itemLen+keys.Len+1] = byte(u >> 16)
			data[i*itemLen+keys.Len+2] = byte(u >> 8)
			data[i*itemLen+keys.Len+3] = byte(u >> 0)
		}
		return data
	})
}
func LoadReverseIndex(file string) (ReverseIndex, error) {
	return indexdatabase.Load(file, func(data []byte) ([]KeyFloat32, error) {
		const itemLen = keys.Len + 4
		if len(data)%itemLen != 0 {
			return nil, fmt.Errorf("Expected len data multiple of %d, get: %d", itemLen, len(data))
		}

		slice := make([]KeyFloat32, len(data)/itemLen)
		for i := range slice {
			key := keys.Key{}
			copy(key[:], data[i*itemLen:])

			p := i*itemLen + keys.Len
			u := 0 |
				uint32(data[p+0])<<24 |
				uint32(data[p+1])<<16 |
				uint32(data[p+2])<<8 |
				uint32(data[p+3])<<0

			slice[i] = KeyFloat32{key, math.Float32frombits(u)}
		}

		return slice, nil
	})
}
