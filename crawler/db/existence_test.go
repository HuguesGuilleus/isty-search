package db

import (
	"crypto/rand"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"testing"
)

var keys = func() (keys [1024]Key) {
	for i := range keys {
		rand.Read(keys[i][:])
	}
	return
}()

func TestExistenceMap(t *testing.T) {
	db := OpenExistenceMap()
	testExistenceAddExistClose(t, db)
}

func TestExistenceFile(t *testing.T) {
	dbPath := "existence.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := OpenExistenceFile(dbPath)
	assert.NoError(t, err)
	testExistenceAddExistClose(t, db)

	expectedMap := make(map[Key]bool, len(keys))
	for _, key := range keys {
		expectedMap[key] = true
	}
	loadedKeys, err := LoadExistenceFile(dbPath)
	assert.NoError(t, err)
	assert.Equal(t, expectedMap, loadedKeys)
}

func testExistenceAddExistClose(t *testing.T, db Existence) {
	for i := range keys {
		assert.NoError(t, db.Add(keys[i]))
	}
	for _, k := range keys {
		assert.True(t, db.Exist(k), k)
	}

	assert.NoError(t, db.Close())
}

func TestExistenceFileFilter(t *testing.T) {
	dbPath := "existence.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := OpenExistenceFile(dbPath)
	defer db.Close()
	assert.NoError(t, err)
	testExistenceFilter(t, db)
}

func TestExistenceMapFilter(t *testing.T) {
	testExistenceFilter(t, OpenExistenceMap())
}

func testExistenceFilter(t *testing.T, db Existence) {
	assert.NoError(t, db.Add(NewURLKey(googleURL)))

	rootGoogleURL, err := googleURL.Parse("/")
	assert.NoError(t, err)

	received := db.Filter([]*url.URL{
		googleURL,
		rootGoogleURL,
		rootGoogleURL,
		googleURL,
	})
	assert.Equal(t, []*url.URL{rootGoogleURL}, received)
}
