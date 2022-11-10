package crawler

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/HuguesGuilleus/isty-search/crawler/robotstxt"
	"net/url"
	"path/filepath"
	"time"
)

type Page struct {
	URL  url.URL
	Time time.Time

	// Content, on of the following filed.
	Error    string
	Html     *htmlnode.Root
	Robots   *robotstxt.File
	Redirect *url.URL
}

// A composite DB with sub subdatabase for specific usage.
type DB struct {
	KeyValueDB db.KeyValueDB[Page]
	URLsDB     *db.URLsDB
}

func OpenDB(root string) (*DB, []*url.URL, error) {
	keyValueDB := db.OpenKeyValueDB[Page](filepath.Join(root, "object"))

	urlsDB, urls, err := db.OpenURLsDB(root, 2048)
	if err != nil {
		return nil, nil, err
	}

	return &DB{
		KeyValueDB: keyValueDB,
		URLsDB:     urlsDB,
	}, urls, nil
}

func (database *DB) Close() error {
	if err := database.URLsDB.Close(); err != nil {
		return fmt.Errorf("Close URLsDB: %w", err)
	}
	database.URLsDB = nil

	return nil
}

// Ban a url for a specific reason.
func (database *DB) ban(u *url.URL, reason string) {
	database.save(u, &Page{Error: reason})
}

// Save a page in the (existence and object) DB.
func (database *DB) save(u *url.URL, page *Page) {
	key := db.NewURLKey(u)

	page.Time = time.Now().UTC()
	page.URL = *u

	if page.Robots == nil {
		database.URLsDB.Store(key, page.Error != "")
	}
	database.KeyValueDB.Store(key, page)
}
