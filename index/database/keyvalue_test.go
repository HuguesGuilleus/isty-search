package indexdatabase

import (
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestKeyValue(t *testing.T) {
	defer os.Remove("_kv.bin")

	m := map[crawldatabase.Key]Bytes{
		crawldatabase.NewKeyString("a"): {1, 2, 3},
		crawldatabase.NewKeyString("b"): {4, 5, 6},
		crawldatabase.NewKeyString("c"): {7, 8, 9},
	}

	assert.NoError(t, Store("_kv.bin", m))

	newMap, err := Load[Bytes]("_kv.bin", UnmarshalBytes)
	assert.NoError(t, err)
	assert.Equal(t, m, newMap)
}
