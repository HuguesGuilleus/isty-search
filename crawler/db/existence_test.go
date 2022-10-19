package db

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
)

func TestExistence(t *testing.T) {
	dbPath := "existence.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := OpenExistence(dbPath)
	assert.NoError(t, err)

	var keys [1024]Key
	for i := range keys {
		rand.Read(keys[i][:])
		assert.NoError(t, db.Add(keys[i]))
	}

	for _, k := range keys {
		assert.True(t, db.Exist(k), k)
	}

	assert.NoError(t, db.Close())

	i := 0
	err = ReadExistence(dbPath, func(key Key) {
		assert.Equal(t, keys[i], key, i)
		i++
	})
	assert.NoError(t, err)

	assert.Equal(t, len(keys), i)
}

func TestExistenceFilter(t *testing.T) {
	dbPath := "existence.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := OpenExistence(dbPath)
	defer db.Close()
	assert.NoError(t, err)
	assert.NoError(t, db.Add(NewURLKey(googleURL)))

	rootGoogleURL, err := googleURL.Parse("/")
	assert.NoError(t, err)

	received := db.Filter([]*url.URL{
		googleURL,
		rootGoogleURL,
		googleURL,
	})
	assert.Equal(t, []*url.URL{rootGoogleURL}, received)
}
