package crawldatabase

import (
	"bytes"
	"crypto/sha256"
	"github.com/stretchr/testify/assert"
	"io"
	"strconv"
	"testing"
)

func writeElasticMetavalue(key Key, meta metavalue, w io.Writer) error {
	bytes := [keyMetavalueLen]byte{}
	copy(bytes[:], key[:])
	bytes[KeyLen] = meta.Type

	switch meta.Type {
	case TypeNothing, TypeKnow:
		_, err := w.Write(bytes[:KeyLen+1])
		return err
	}

	bytes[KeyLen+1] = byte(meta.Time >> 48)
	bytes[KeyLen+2] = byte(meta.Time >> 40)
	bytes[KeyLen+3] = byte(meta.Time >> 32)
	bytes[KeyLen+4] = byte(meta.Time >> 24)
	bytes[KeyLen+5] = byte(meta.Time >> 16)
	bytes[KeyLen+6] = byte(meta.Time >> 8)
	bytes[KeyLen+7] = byte(meta.Time >> 0)

	if meta.Type >= TypeError {
		_, err := w.Write(bytes[:40])
		return err
	}

	switch meta.Type {
	case TypeRedirect:
		bytes[KeyLen] = TypeRedirect
		copy(bytes[KeyLen+8:], meta.Hash[:])
	default:
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

		copy(bytes[52:], meta.Hash[12:])
	}
	_, err := w.Write(bytes[:])
	return err
}

func TestWriteElasticMetavalueSimple(t *testing.T) {
	fn := func(t *testing.T, byteType byte) {
		buff := bytes.Buffer{}
		err := writeElasticMetavalue(NewKeyURL(googleHowURL), metavalue{Type: byteType}, &buff)
		assert.NoError(t, err)

		assert.Equal(t, []byte{
			// Key
			0x3d, 0xd2, 0x98, 0x19, 0x98, 0x42, 0x30, 0x88, 0x39, 0xe8, 0xf2, 0xd7, 0xe8, 0xf6, 0x58, 0x51, 0x54, 0xe3, 0xce, 0x49, 0xe7, 0x7c, 0xcc, 0x45, 0x34, 0x0a, 0x5b, 0x06, 0x4e, 0xac, 0xdd, 0xfe,
			// The type
			byteType,
		}, buff.Bytes())
	}

	t.Run("nothing", func(t *testing.T) { fn(t, TypeNothing) })
	t.Run("known", func(t *testing.T) { fn(t, TypeKnow) })
}
func TestWriteElasticMetavalueRedirect(t *testing.T) {
	buff := bytes.Buffer{}
	err := writeElasticMetavalue(NewKeyURL(googleHowURL), metavalue{
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
func TestWriteElasticMetavalueFile(t *testing.T) {
	buff := bytes.Buffer{}
	err := writeElasticMetavalue(NewKeyURL(googleHowURL), metavalue{
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
func TestWriteElasticMetavalueError(t *testing.T) {
	buff := bytes.Buffer{}
	err := writeElasticMetavalue(NewKeyURL(googleHowURL), metavalue{
		Type: TypeErrorNetwork,
		Time: 1671022548,
	}, &buff)
	assert.NoError(t, err)

	assert.Equal(t, []byte{
		// Key
		0x3d, 0xd2, 0x98, 0x19, 0x98, 0x42, 0x30, 0x88, 0x39, 0xe8, 0xf2, 0xd7, 0xe8, 0xf6, 0x58, 0x51, 0x54, 0xe3, 0xce, 0x49, 0xe7, 0x7c, 0xcc, 0x45, 0x34, 0x0a, 0x5b, 0x06, 0x4e, 0xac, 0xdd, 0xfe,
		// Type
		128,
		// Time
		0, 0, 0, 0x63, 0x99, 0xc7, 0xd4,
	}, buff.Bytes())
}

/* LOADER */

// Load many meta and keys
func loadElasticMetavalue(bytes []byte) map[Key]metavalue {
	mapMeta := make(map[Key]metavalue, len(bytes)/(KeyLen+1))

	for i := 0; i < len(bytes); {
		key := Key{}
		copy(key[:], bytes[i:])

		metaType := bytes[i+KeyLen]
		if metaType == TypeNothing {
			delete(mapMeta, key)
			i += KeyLen + 1
			continue
		} else if metaType == TypeKnow {
			mapMeta[key] = metavalue{Type: TypeKnow}
			i += KeyLen + 1
			continue
		}

		// TODO: test s'il manque des valeurs à la fin.

		meta := metavalue{Type: metaType}
		meta.Time = 0 |
			int64(bytes[i+KeyLen+1])<<48 |
			int64(bytes[i+KeyLen+2])<<40 |
			int64(bytes[i+KeyLen+3])<<32 |
			int64(bytes[i+KeyLen+4])<<24 |
			int64(bytes[i+KeyLen+5])<<16 |
			int64(bytes[i+KeyLen+6])<<8 |
			int64(bytes[i+KeyLen+7])<<0

		if meta.Type >= TypeError {
			mapMeta[key] = meta
			i += KeyLen + 1 + 7
			continue
		}

		switch meta.Type {
		case TypeRedirect:
			copy(meta.Hash[:], bytes[i+KeyLen+8:])
		default:
			if meta.Type < TypeError { // It's a file
				meta.Position = 0 |
					int64(bytes[i+KeyLen+8])<<56 |
					int64(bytes[i+KeyLen+9])<<48 |
					int64(bytes[i+KeyLen+10])<<40 |
					int64(bytes[i+KeyLen+11])<<32 |
					int64(bytes[i+KeyLen+12])<<24 |
					int64(bytes[i+KeyLen+13])<<16 |
					int64(bytes[i+KeyLen+14])<<8 |
					int64(bytes[i+KeyLen+15])<<0
				meta.Length = 0 |
					int32(bytes[i+KeyLen+16])<<24 |
					int32(bytes[i+KeyLen+17])<<16 |
					int32(bytes[i+KeyLen+18])<<8 |
					int32(bytes[i+KeyLen+19])<<0
				copy(meta.Hash[12:], bytes[i+KeyLen+20:])
			}
		}
		mapMeta[key] = meta

		i += keyMetavalueLen
	}

	return mapMeta
}

func TestLoadElasticMetavalue(t *testing.T) {
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
		expectedMetavalue[key] = m

		assert.NoError(t, writeElasticMetavalue(key, m, &buff))
	}
	delete(expectedMetavalue, NewKeyString("https://www.google.com/0"))

	receivedMap := loadElasticMetavalue(buff.Bytes())
	assert.Equal(t, len(expectedMetavalue), len(receivedMap))

	for key := range receivedMap {
		t.Log(key)
	}

	for key, meta := range expectedMetavalue {
		assert.Equal(t, meta, receivedMap[key], key)
	}
}

var benchmarkKey = NewKeyString("https://www.google.com/")

func BenchmarkMEtavalue(b *testing.B) {
	bench := func(b *testing.B, writer func(Key, metavalue, io.Writer) error, loader func([]byte) map[Key]metavalue) {
		buff := bytes.Buffer{}
		buff.Grow(b.N * 256 * keyMetavalueLen)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buff.Reset()
			for i := 0; i < 11228; i++ {
				writer(benchmarkKey, metavalue{
					Type:     TypeFile,
					Time:     0x11223344556677,
					Hash:     benchmarkKey,
					Position: 0x1122334455667788,
					Length:   0x11223344,
				}, &buff)
			}
			for i := 0; i < 57208; i++ {
				writer(benchmarkKey, metavalue{
					Type:     TypeError,
					Time:     0x11223344556677,
					Hash:     benchmarkKey,
					Position: 0x1122334455667788,
					Length:   0x11223344,
				}, &buff)
			}
			for i := 0; i < 9181; i++ {
				writer(benchmarkKey, metavalue{
					Type:     TypeKnow,
					Time:     0x11223344556677,
					Hash:     benchmarkKey,
					Position: 0x1122334455667788,
					Length:   0x11223344,
				}, &buff)
			}
			// for i := 0; i < 256; i++ {
			// 	writer(benchmarkKey, metavalue{
			// 		Type:     byte(i),
			// 		Time:     0x11223344556677,
			// 		Hash:     benchmarkKey,
			// 		Position: 0x1122334455667788,
			// 		Length:   0x11223344,
			// 	}, &buff)
			// }
			loader(buff.Bytes())
		}
	}

	b.Run("static", func(b *testing.B) {
		bench(b, writeMetavalue, loadMetavalue)
	})
	b.Run("elastic", func(b *testing.B) {
		bench(b, writeElasticMetavalue, loadElasticMetavalue)
	})
}
