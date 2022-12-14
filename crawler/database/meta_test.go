package crawldatabase

import (
	"bytes"
	"crypto/sha256"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

var (
	googleHowURL  = common.ParseURL("https://www.google.com/search/howsearchworks/?fg=1")
	googleRootURL = common.ParseURL("https://www.google.com/")
)

func TestKeyPath(t *testing.T) {
	assert.Equal(t,
		"3dd298199842308839e8f2d7e8f6585154e3ce49e77ccc45340a5b064eacddfe",
		NewKeyURL(googleHowURL).String(),
	)
}

func TestWriteMetavalueSimple(t *testing.T) {
	fn := func(t *testing.T, byteType byte) {
		buff := bytes.Buffer{}
		err := writeMetavalue(NewKeyURL(googleHowURL), metavalue{Type: byteType}, &buff)
		assert.NoError(t, err)

		expected := [72]byte{
			// Key
			0x3d, 0xd2, 0x98, 0x19, 0x98, 0x42, 0x30, 0x88, 0x39, 0xe8, 0xf2, 0xd7, 0xe8, 0xf6, 0x58, 0x51, 0x54, 0xe3, 0xce, 0x49, 0xe7, 0x7c, 0xcc, 0x45, 0x34, 0x0a, 0x5b, 0x06, 0x4e, 0xac, 0xdd, 0xfe,
			// The type
			byteType,
			// Zeros for remains
		}
		assert.Equal(t, expected[:], buff.Bytes())
	}

	t.Run("nothing", func(t *testing.T) { fn(t, TypeNothing) })
	t.Run("known", func(t *testing.T) { fn(t, TypeKnow) })
}

func TestWriteMetavalueRedirect(t *testing.T) {
	buff := bytes.Buffer{}
	err := writeMetavalue(NewKeyURL(googleHowURL), metavalue{
		Type: TypeRedirect,
		Time: 1671022548,
		Hash: NewKeyURL(googleRootURL),
	}, &buff)
	assert.NoError(t, err)

	expected := [72]byte{
		// Key
		0x3d, 0xd2, 0x98, 0x19, 0x98, 0x42, 0x30, 0x88, 0x39, 0xe8, 0xf2, 0xd7, 0xe8, 0xf6, 0x58, 0x51, 0x54, 0xe3, 0xce, 0x49, 0xe7, 0x7c, 0xcc, 0x45, 0x34, 0x0a, 0x5b, 0x06, 0x4e, 0xac, 0xdd, 0xfe,
		// Type
		2,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
		// URL
		0xd0, 0xe1, 0x96, 0xa0, 0xc2, 0x5d, 0x35, 0xdd, 0xa, 0x84, 0x59, 0x3c, 0xba, 0xe0, 0xf3, 0x83, 0x33, 0xaa, 0x58, 0x52, 0x99, 0x36, 0x44, 0x4e, 0xa2, 0x64, 0x53, 0xea, 0xb2, 0x8d, 0xfc, 0x86,
	}
	assert.Equal(t, expected[:], buff.Bytes())
}

func TestWriteMetavalueFile(t *testing.T) {
	buff := bytes.Buffer{}
	err := writeMetavalue(NewKeyURL(googleHowURL), metavalue{
		Type:     TypeFileHTML,
		Time:     1671022548,
		Position: 0x1122334455667788,
		Length:   0x11223344,
		Hash:     NewKeyURL(googleRootURL),
	}, &buff)
	assert.NoError(t, err)

	expected := [72]byte{
		// Key
		0x3d, 0xd2, 0x98, 0x19, 0x98, 0x42, 0x30, 0x88, 0x39, 0xe8, 0xf2, 0xd7, 0xe8, 0xf6, 0x58, 0x51, 0x54, 0xe3, 0xce, 0x49, 0xe7, 0x7c, 0xcc, 0x45, 0x34, 0x0a, 0x5b, 0x06, 0x4e, 0xac, 0xdd, 0xfe,
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
	}
	assert.Equal(t, expected[:], buff.Bytes())
}

func TestWriteMetavalueError(t *testing.T) {
	buff := bytes.Buffer{}
	err := writeMetavalue(NewKeyURL(googleHowURL), metavalue{
		Type: TypeErrorNetwork,
		Time: 1671022548,
	}, &buff)
	assert.NoError(t, err)

	expected := [72]byte{
		// Key
		0x3d, 0xd2, 0x98, 0x19, 0x98, 0x42, 0x30, 0x88, 0x39, 0xe8, 0xf2, 0xd7, 0xe8, 0xf6, 0x58, 0x51, 0x54, 0xe3, 0xce, 0x49, 0xe7, 0x7c, 0xcc, 0x45, 0x34, 0x0a, 0x5b, 0x06, 0x4e, 0xac, 0xdd, 0xfe,
		// Type
		128,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
		// Zero remain
	}
	assert.Equal(t, expected[:], buff.Bytes())
}

func TestLoadMetavalue(t *testing.T) {
	hash := sha256.Sum256([]byte("hello world"))
	for i := 0; i < 12; i++ {
		hash[i] = 0
	}

	metavalueOrigin := []metavalue{
		metavalue{Type: TypeNothing},
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
		t.Log(i, key)

		err := writeMetavalue(key, m, &buff)
		assert.NoError(t, err)

		expectedMetavalue[key] = m
	}

	receivedMap := loadMetavalue(buff.Bytes())
	assert.Equal(t, len(expectedMetavalue), len(receivedMap))
	for key, meta := range expectedMetavalue {
		assert.Equal(t, meta, receivedMap[key], key)
	}
}
