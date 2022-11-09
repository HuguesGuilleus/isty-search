package db

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"path/filepath"
)

const KeyLen = sha256.Size

// One objet DB key.
type Key [KeyLen]byte

func NewURLKey(u *url.URL) Key {
	return NewStringKey(u.String())
}

func NewStringKey(s string) Key {
	return Key(sha256.Sum256([]byte(s)))
}

// Retunrn the value in hexadecimal
func (key Key) String() string {
	return hex.EncodeToString(key[:])
}

// From cleaned base path of the db, return "base/xx/xx/xx/xx...xx".
func (key Key) path(base string) string {
	l := len(base)
	p := make([]byte, len(base)+4+2*sha256.Size)

	p[l+0] = filepath.Separator
	p[l+3] = filepath.Separator
	p[l+6] = filepath.Separator
	p[l+9] = filepath.Separator

	copy(p, base)
	hex.Encode(p[l+1:], key[0:1])
	hex.Encode(p[l+4:], key[1:2])
	hex.Encode(p[l+7:], key[2:3])
	hex.Encode(p[l+10:], key[3:])

	return string(p)
}

// From cleaned base path of the db, return "base/xx/xx/xx" for the directory
// where store the file.
func (key Key) dir(base string) string {
	l := len(base)
	p := make([]byte, len(base)+3+2*3)

	p[l+0] = filepath.Separator
	p[l+3] = filepath.Separator
	p[l+6] = filepath.Separator

	copy(p, base)
	hex.Encode(p[l+1:], key[0:1])
	hex.Encode(p[l+4:], key[1:2])
	hex.Encode(p[l+7:], key[2:3])

	return string(p)
}
