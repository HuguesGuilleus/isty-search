package crawldatabase

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKey(t *testing.T) {
	assert.Equal(t,
		"3dd298199842308839e8f2d7e8f6585154e3ce49e77ccc45340a5b064eacddfe",
		NewKeyURL(googleHowURL).String(),
	)
}
