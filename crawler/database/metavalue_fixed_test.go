package crawldatabase

// This file has writeStaticMetavalue and loadStaticMetavalue, each entry have
// always 72 byte.
// The usage of this function is to benchmark with the elastic version.

import (
	"bytes"
	"github.com/HuguesGuilleus/isty-search/keys"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

// Save into w, the key and the metavalue.
// The size of each entry is the same.
func writeFixedMetavalue(key keys.Key, meta metavalue, w io.Writer) error {
	bytes := [keyMetavalueLen]byte{}
	copy(bytes[:], key[:])

	switch meta.Type {
	case TypeNothing, TypeKnow:
		meta.Time = 0
	case TypeRedirect:
		copy(bytes[40:], meta.Hash[:])
	default:
		if meta.Type < TypeError { // file type
			bytes[40] = byte(meta.Position >> 56)
			bytes[41] = byte(meta.Position >> 48)
			bytes[42] = byte(meta.Position >> 40)
			bytes[43] = byte(meta.Position >> 32)
			bytes[44] = byte(meta.Position >> 24)
			bytes[45] = byte(meta.Position >> 16)
			bytes[46] = byte(meta.Position >> 8)
			bytes[47] = byte(meta.Position >> 0)

			bytes[48] = byte(meta.Length >> 24)
			bytes[49] = byte(meta.Length >> 16)
			bytes[50] = byte(meta.Length >> 8)
			bytes[51] = byte(meta.Length >> 0)

			copy(bytes[52:72], meta.Hash[12:32])
		}
	}

	bytes[32] = meta.Type
	bytes[33] = byte(meta.Time >> 48)
	bytes[34] = byte(meta.Time >> 40)
	bytes[35] = byte(meta.Time >> 32)
	bytes[36] = byte(meta.Time >> 24)
	bytes[37] = byte(meta.Time >> 16)
	bytes[38] = byte(meta.Time >> 8)
	bytes[39] = byte(meta.Time >> 0)

	_, err := w.Write(bytes[:])
	return err
}
func TestWriteFixedMetavalue(t *testing.T) {
	// A function to test the writer.
	// The key is appended to the expected bytes.
	test := func(name string, m metavalue, expected [40]byte) {
		t.Run(name, func(t *testing.T) {
			key := keys.NewURL(googleHowURL)
			buff := bytes.Buffer{}
			assert.NoError(t, writeFixedMetavalue(key, m, &buff))
			assert.Equal(t, append(key[:], expected[:]...), buff.Bytes())
		})
	}

	test("nothing", metavalue{Type: TypeNothing}, [40]byte{TypeNothing})
	test("known", metavalue{Type: TypeKnow}, [40]byte{TypeKnow})

	test("redirect", metavalue{
		Type: TypeRedirect,
		Time: 0x00_0000_6399_c7d4,
		Hash: keys.NewURL(googleRootURL),
	}, [40]byte{
		// Type
		2,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
		// URL
		0xd0, 0xe1, 0x96, 0xa0, 0xc2, 0x5d, 0x35, 0xdd, 0xa, 0x84, 0x59, 0x3c, 0xba, 0xe0, 0xf3, 0x83, 0x33, 0xaa, 0x58, 0x52, 0x99, 0x36, 0x44, 0x4e, 0xa2, 0x64, 0x53, 0xea, 0xb2, 0x8d, 0xfc, 0x86,
	})

	test("file", metavalue{
		Type:     TypeFileHTML,
		Time:     0x00_0000_6399_c7d4,
		Position: 0x1122334455667788,
		Length:   0x11223344,
		Hash:     keys.NewURL(googleRootURL),
	}, [40]byte{
		// Type HTML
		4,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
		// Position
		0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88,
		// Length
		0x11, 0x22, 0x33, 0x44,
		// End of hash of the content
		0xba, 0xe0, 0xf3, 0x83,
		0x33, 0xaa, 0x58, 0x52, 0x99, 0x36, 0x44, 0x4e,
		0xa2, 0x64, 0x53, 0xea, 0xb2, 0x8d, 0xfc, 0x86,
	})

	test("error", metavalue{
		Type: TypeErrorNetwork,
		Time: 0x00_0000_6399_c7d4,
	}, [40]byte{
		// Type
		128,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
	})
}

// Load many meta and keys.
// The size of each entry is the same.
func loadFixedMetavalue(bytes []byte) map[keys.Key]metavalue {
	mapMeta := make(map[keys.Key]metavalue, len(bytes)/keyMetavalueLen)

	for i := 0; i < len(bytes); i += keyMetavalueLen {
		key := keys.Key{}
		copy(key[:], bytes[i:])

		meta := metavalue{}

		meta.Time = 0 |
			int64(bytes[i+keys.Len+1])<<48 |
			int64(bytes[i+keys.Len+2])<<40 |
			int64(bytes[i+keys.Len+3])<<32 |
			int64(bytes[i+keys.Len+4])<<24 |
			int64(bytes[i+keys.Len+5])<<16 |
			int64(bytes[i+keys.Len+6])<<8 |
			int64(bytes[i+keys.Len+7])<<0

		switch meta.Type = bytes[i+keys.Len]; meta.Type {
		case TypeNothing:
			delete(mapMeta, key)
			continue
		case TypeKnow:
			meta.Time = 0
		case TypeRedirect:
			copy(meta.Hash[:], bytes[i+keys.Len+8:])
		default:
			if meta.Type < TypeError { // It's a file
				meta.Position = 0 |
					int64(bytes[i+keys.Len+8])<<56 |
					int64(bytes[i+keys.Len+9])<<48 |
					int64(bytes[i+keys.Len+10])<<40 |
					int64(bytes[i+keys.Len+11])<<32 |
					int64(bytes[i+keys.Len+12])<<24 |
					int64(bytes[i+keys.Len+13])<<16 |
					int64(bytes[i+keys.Len+14])<<8 |
					int64(bytes[i+keys.Len+15])<<0
				meta.Length = 0 |
					int32(bytes[i+keys.Len+16])<<24 |
					int32(bytes[i+keys.Len+17])<<16 |
					int32(bytes[i+keys.Len+18])<<8 |
					int32(bytes[i+keys.Len+19])<<0
				copy(meta.Hash[12:], bytes[i+keys.Len+20:])
			}
		}

		mapMeta[key] = meta
	}

	return mapMeta
}
func TestLoadMetavalue(t *testing.T) {
	testLoaderMetavalue(t, writeFixedMetavalue, loadFixedMetavalue)
}
