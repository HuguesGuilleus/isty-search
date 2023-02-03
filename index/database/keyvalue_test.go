package indexdatabase

import (
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestKeyValue(t *testing.T) {
	defer os.Remove("_kv.bin")

	m := map[keys.Key][]byte{
		keys.Key{0xA}: {1, 2, 3},
		keys.Key{0xB}: {4, 5, 6},
		keys.Key{0xD}: {7, 8, 9},
	}

	assert.NoError(t, Store("_kv.bin", m, func(data []byte) []byte { return data }))

	newMap, err := Load("_kv.bin", func(data []byte) ([]byte, error) { return data, nil })
	assert.NoError(t, err)
	assert.EqualValues(t, m, newMap)
}
