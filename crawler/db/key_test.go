package database

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestKeyPath(t *testing.T) {
	u, err := url.Parse("https://www.google.com/search/howsearchworks/?fg=1")
	if err != nil {
		assert.NoError(t, err)
	}

	assert.Equal(t, "base/3d/d2/98/199842308839e8f2d7e8f6585154e3ce49e77ccc45340a5b064eacddfe", NewURLKey(u).path("base"))
}
