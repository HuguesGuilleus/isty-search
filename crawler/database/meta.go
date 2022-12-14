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

	TypeError        byte = 128
	TypeErrorNetwork byte = 128
)

const (
	// The length in byte of the key and the meta value .
	KeyMetaLen = 72
	// A bits mask to get the time.
	timeMask = 1<<56 - 1
)

// The meta value, different type: nothing | known | redirect | file | error
type metavalue struct {
	Type     byte
	Time     int64
	Hash     Key
	Position int64
	Length   int32
}

func writeMetaValue(key Key, meta metavalue, w io.Writer) error {
	bytes := [72]byte{}
	copy(bytes[:], key[:])

	switch meta.Type {
	case TypeNothing:
		meta.Time = 0
	case TypeKnow:
		meta.Time = 0
	case TypeRedirect:
		bytes[KeyLen] = TypeRedirect
		copy(bytes[KeyLen+8:], meta.Hash[:])
	default:
		if meta.Type < TypeError { // a file
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
	}

	bytes[KeyLen] = meta.Type
	bytes[KeyLen+1] = byte(meta.Time >> 48)
	bytes[KeyLen+2] = byte(meta.Time >> 40)
	bytes[KeyLen+3] = byte(meta.Time >> 32)
	bytes[KeyLen+4] = byte(meta.Time >> 24)
	bytes[KeyLen+5] = byte(meta.Time >> 16)
	bytes[KeyLen+6] = byte(meta.Time >> 8)
	bytes[KeyLen+7] = byte(meta.Time >> 0)

	_, err := w.Write(bytes[:])
	return err
}
