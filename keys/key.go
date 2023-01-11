package keys

import (
	"crypto/sha256"
	"encoding/hex"
	"net/url"
)

const Len = sha256.Size

// One objet DB key.
type Key [Len]byte

// Hash the URL to create a Key
func NewURL(u *url.URL) Key { return NewString(u.String()) }

// Hash the string to create a Key.
func NewString(s string) Key { return Key(sha256.Sum256([]byte(s))) }

// Return the value in hexadecimal
func (key Key) String() string { return hex.EncodeToString(key[:]) }
