package crawldatabase

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
)

const KeyLen = sha256.Size

// One objet DB key.
type Key [KeyLen]byte

func NewKeyURL(u *url.URL) Key  { return NewKeyString(u.String()) }
func NewKeyString(s string) Key { return Key(sha256.Sum256([]byte(s))) }

// Return the value in hexadecimal
func (key Key) String() string {
	return hex.EncodeToString(key[:])
}
