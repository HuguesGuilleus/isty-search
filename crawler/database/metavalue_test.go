package crawldatabase

import (
	"bytes"
	"crypto/sha256"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/stretchr/testify/assert"
	"io"
	"strconv"
	"testing"
)

var (
	googleHowURL  = common.ParseURL("https://www.google.com/search/howsearchworks/?fg=1")
	googleRootURL = common.ParseURL("https://www.google.com/")
)

func TestWriteElasticMetavalue(t *testing.T) {
	// A function to test the writer.
	// The key is appended to the expected bytes.
	testWriteMetavalue := func(name string, m metavalue, expected []byte) {
		t.Run(name, func(t *testing.T) {
			key := NewKeyURL(googleHowURL)
			buff := bytes.Buffer{}
			assert.NoError(t, writeElasticMetavalue(key, m, &buff))
			assert.Equal(t, append(key[:], expected...), buff.Bytes())
		})
	}

	testWriteMetavalue("nothing", metavalue{
		Type: TypeNothing,
	}, []byte{TypeNothing})

	testWriteMetavalue("know", metavalue{
		Type: TypeKnow,
	}, []byte{TypeKnow})

	testWriteMetavalue("redirect", metavalue{
		Type: TypeRedirect,
		Time: 0x00_0000_6399_c7d4,
		Hash: NewKeyURL(googleRootURL),
	}, []byte{
		// Type
		2,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
		// URL
		0xd0, 0xe1, 0x96, 0xa0, 0xc2, 0x5d, 0x35, 0xdd, 0xa, 0x84, 0x59, 0x3c, 0xba, 0xe0, 0xf3, 0x83, 0x33, 0xaa, 0x58, 0x52, 0x99, 0x36, 0x44, 0x4e, 0xa2, 0x64, 0x53, 0xea, 0xb2, 0x8d, 0xfc, 0x86,
	})

	testWriteMetavalue("file", metavalue{
		Type:     TypeFileHTML,
		Time:     0x00_0000_6399_c7d4,
		Position: 0x1122334455667788,
		Length:   0x11223344,
		Hash:     NewKeyURL(googleRootURL),
	}, []byte{
		// Type HTML
		4,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
		// Position
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
		// Length
		0x11, 0x22, 0x33, 0x44,
		// end of hash of the content
		0xba, 0xe0, 0xf3, 0x83,
		0x33, 0xaa, 0x58, 0x52, 0x99, 0x36, 0x44, 0x4e,
		0xa2, 0x64, 0x53, 0xea, 0xb2, 0x8d, 0xfc, 0x86,
	})

	testWriteMetavalue("error", metavalue{
		Type: TypeErrorNetwork,
		Time: 0x00_0000_6399_c7d4,
	}, []byte{
		// Type
		128,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
	})
}

func testLoaderMetavalue(t *testing.T, writer func(Key, metavalue, io.Writer) error, loader func([]byte) map[Key]metavalue) {
	hash := sha256.Sum256([]byte("hello world"))
	for i := 0; i < 12; i++ {
		hash[i] = 0
	}
	metavalueOrigin := []metavalue{
		metavalue{Type: TypeKnow}, // wil be deleted
		metavalue{Type: TypeKnow},
		metavalue{
			Type: TypeRedirect,
			Time: 1671022548,
			Hash: NewKeyURL(googleRootURL),
		},
		metavalue{
			Type:     TypeFileHTML,
			Time:     1671022548,
			Position: 0x1122334455667788,
			Length:   0x11223344,
			Hash:     hash,
		},
		metavalue{
			Type: TypeErrorNetwork,
			Time: 1671022548,
		},
	}

	expectedMetavalue := make(map[Key]metavalue, len(metavalueOrigin))
	buff := bytes.Buffer{}
	for i, m := range metavalueOrigin {
		key := NewKeyString("https://www.google.com/" + strconv.Itoa(i))
		expectedMetavalue[key] = m

		assert.NoError(t, writer(key, m, &buff))
	}

	key0 := NewKeyString("https://www.google.com/0")
	assert.NoError(t, writer(key0, metavalue{Type: TypeNothing}, &buff))
	delete(expectedMetavalue, key0)

	receivedMap := loader(buff.Bytes())
	assert.Equal(t, len(expectedMetavalue), len(receivedMap))
	for key, meta := range expectedMetavalue {
		assert.Equal(t, meta, receivedMap[key], key)
	}
}
func TestLoadElasticMetavalue(t *testing.T) {
	testLoaderMetavalue(t, writeElasticMetavalue, loadElasticMetavalue)
}

func BenchmarkMetavalue(b *testing.B) {
	key := NewKeyURL(googleRootURL)
	buff := bytes.Buffer{}
	buff.Grow(256 * keyMetavalueLen)

	write := func(writer func(Key, metavalue, io.Writer) error) {
		buff.Reset()
		// This value come from the first crawl.
		for i := 0; i < 11228; i++ {
			writer(key, metavalue{
				Type:     TypeFile,
				Time:     0x11223344556677,
				Hash:     key,
				Position: 0x1122334455667788,
				Length:   0x11223344,
			}, &buff)
		}
		for i := 0; i < 57208; i++ {
			writer(key, metavalue{
				Type: TypeError,
				Time: 0x11223344556677,
			}, &buff)
		}
		for i := 0; i < 9181; i++ {
			writer(key, metavalue{Type: TypeKnow}, &buff)
		}
	}

	write(writeFixedMetavalue)
	b.Log("fixed data len:  ", buff.Len())
	write(writeElasticMetavalue)
	b.Log("elastic data len:", buff.Len())

	bench := func(name string, writer func(Key, metavalue, io.Writer) error, loader func([]byte) map[Key]metavalue) {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				write(writer)
				loader(buff.Bytes())
			}
		})
	}

	bench("fixed", writeFixedMetavalue, loadFixedMetavalue)
	bench("elastic", writeElasticMetavalue, loadElasticMetavalue)
}
