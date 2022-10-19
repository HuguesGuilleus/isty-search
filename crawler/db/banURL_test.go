package db

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
)

func TestBanURL(t *testing.T) {
	filePath := "ban.txt"
	os.Remove(filePath)
	defer os.Remove(filePath)

	db, err := OpenBanURL(filePath)
	assert.NoError(t, err)

	db.Add(googleURL)

	i := 0
	err = ReadBanURL(filePath, func(u *url.URL) {
		assert.Equal(t, googleURL, u)
		i++
	})
	assert.Equal(t, 1, i)
	assert.NoError(t, err)
}
