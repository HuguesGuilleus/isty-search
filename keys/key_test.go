package keys

import (
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKey(t *testing.T) {
	assert.Equal(t,
		"3dd298199842308839e8f2d7e8f6585154e3ce49e77ccc45340a5b064eacddfe",
		NewURL(common.ParseURL("https://www.google.com/search/howsearchworks/?fg=1")).String(),
	)
}

func TestCompare(t *testing.T) {
	assert.True(t, (&Key{0, 1, 3}).Less(&Key{0, 1, 5}))
	assert.False(t, (&Key{}).Less(&Key{}))
}
