package crawldatabase

import (
	"io"
)

const (
	TypeNothing  byte = 0
	TypeKnow     byte = 1
	TypeRedirect byte = 2

	TypeFile        byte = 3
	TypeFileRobots  byte = 3
	TypeFileHTML    byte = 4
	TypeFileRSS     byte = 5
	TypeFileSitemap byte = 6
	TypeFileFavicon byte = 7

	TypeError           byte = 128
	TypeErrorNetwork    byte = 128
	TypeErrorParsing    byte = 129
	TypeErrorFilterURL  byte = 130
	TypeErrorFilterPage byte = 131
)

// The maximum length of the key and the metavalue.
const keyMetavalueLen = 72

type keymetavalue struct {
	key  Key
	meta metavalue
}

// The meta value, different type: nothing | known | redirect | file | error
type metavalue struct {
	Type     byte
	Time     int64
	Hash     Key
	Position int64
	Length   int32
}

func writeElasticMetavalue(key Key, meta metavalue, w io.Writer) error {
	bytes := [keyMetavalueLen]byte{}
	copy(bytes[:], key[:])
	bytes[32] = meta.Type

	switch meta.Type {
	case TypeNothing, TypeKnow:
		_, err := w.Write(bytes[:33])
		return err
	}

	bytes[33] = byte(meta.Time >> 48)
	bytes[34] = byte(meta.Time >> 40)
	bytes[35] = byte(meta.Time >> 32)
	bytes[36] = byte(meta.Time >> 24)
	bytes[37] = byte(meta.Time >> 16)
	bytes[38] = byte(meta.Time >> 8)
	bytes[39] = byte(meta.Time)

	if meta.Type >= TypeError {
		_, err := w.Write(bytes[:40])
		return err
	}

	switch meta.Type {
	case TypeRedirect:
		copy(bytes[40:], meta.Hash[:])
	default: // file
		bytes[40] = byte(meta.Position >> 56)
		bytes[41] = byte(meta.Position >> 48)
		bytes[42] = byte(meta.Position >> 40)
		bytes[43] = byte(meta.Position >> 32)
		bytes[44] = byte(meta.Position >> 24)
		bytes[45] = byte(meta.Position >> 16)
		bytes[46] = byte(meta.Position >> 8)
		bytes[47] = byte(meta.Position)

		bytes[48] = byte(meta.Length >> 24)
		bytes[49] = byte(meta.Length >> 16)
		bytes[50] = byte(meta.Length >> 8)
		bytes[51] = byte(meta.Length)

		copy(bytes[52:], meta.Hash[12:])
	}
	_, err := w.Write(bytes[:])
	return err
}

// Load many meta and keys
func loadElasticMetavalue(bytes []byte) map[Key]metavalue {
	mapMeta := make(map[Key]metavalue, len(bytes)/(KeyLen+1))

	for i := 0; i+32 < len(bytes); {
		key := Key{}
		copy(key[:], bytes[i:])

		metaType := bytes[i+KeyLen]
		if metaType == TypeNothing {
			delete(mapMeta, key)
			i += 33
			continue
		} else if metaType == TypeKnow {
			mapMeta[key] = metavalue{Type: TypeKnow}
			i += 33
			continue
		}

		if i+39 >= len(bytes) {
			break
		}

		meta := metavalue{Type: metaType}
		meta.Time = 0 |
			int64(bytes[i+33])<<48 |
			int64(bytes[i+34])<<40 |
			int64(bytes[i+35])<<32 |
			int64(bytes[i+36])<<24 |
			int64(bytes[i+37])<<16 |
			int64(bytes[i+38])<<8 |
			int64(bytes[i+39])

		if meta.Type >= TypeError {
			mapMeta[key] = meta
			i += 40
			continue
		}

		if i+71 >= len(bytes) {
			break
		}

		switch meta.Type {
		case TypeRedirect:
			copy(meta.Hash[:], bytes[i+40:])
		default:
			if meta.Type < TypeError { // It's a file
				meta.Position = 0 |
					int64(bytes[i+40])<<56 |
					int64(bytes[i+41])<<48 |
					int64(bytes[i+42])<<40 |
					int64(bytes[i+43])<<32 |
					int64(bytes[i+44])<<24 |
					int64(bytes[i+45])<<16 |
					int64(bytes[i+46])<<8 |
					int64(bytes[i+47])<<0
				meta.Length = 0 |
					int32(bytes[i+48])<<24 |
					int32(bytes[i+49])<<16 |
					int32(bytes[i+50])<<8 |
					int32(bytes[i+51])<<0
				copy(meta.Hash[12:], bytes[i+52:])
			}
		}
		mapMeta[key] = meta

		i += 72
	}

	return mapMeta
}
