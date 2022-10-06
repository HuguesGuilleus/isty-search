package db

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"
)

var googleURL = func() *url.URL {
	u, err := url.Parse("https://www.google.com/search/howsearchworks/?fg=1")
	if err != nil {
		panic(err)
	}
	return u
}()

func TestKeyPath(t *testing.T) {
	assert.Equal(t, "base/3d/d2/98/199842308839e8f2d7e8f6585154e3ce49e77ccc45340a5b064eacddfe",
		NewURLKey(googleURL).path("base"))
	assert.Equal(t, "base/3d/d2/98", NewURLKey(googleURL).dir("base"))
}

func TestObjectBD(t *testing.T) {
	// Web use http.Cookie because it's a struct.
	expected := &http.Cookie{
		Name:    "yolo",
		Value:   "swag",
		Expires: time.Date(2022, time.October, 5, 20, 8, 50, 0, time.UTC),
	}
	k := NewURLKey(googleURL)

	db := OpenObjectBD[http.Cookie]("./test_db")
	defer os.RemoveAll("test_db")

	assert.NoError(t, db.Store(k, expected))

	v, err := db.Get(k)
	assert.NoError(t, err)
	assert.Equal(t, expected, v)
}
