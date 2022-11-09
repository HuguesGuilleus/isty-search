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
	Existence  db.Existence
	Found      *db.Found
	Ban        *db.BanURL
}

func OpenDB(root string) (*DB, []*url.URL, error) {
	existence, err := db.OpenExistenceFile(filepath.Join(root, "existence-key.bin"))
	if err != nil {
		return nil, nil, err
	}

	found, urls, err := db.OpenFound(filepath.Join(root, "found.txt"))
	if err != nil {
		return nil, nil, err
	}

	ban, err := db.OpenBanURL(filepath.Join(root, "ban.txt"))
	if err != nil {
		return nil, nil, err
	}

	return &DB{
		KeyValueDB: db.OpenKeyValueDB[Page](filepath.Join(root, "object")),
		Existence:  existence,
		Found:      found,
		Ban:        ban,
	}, urls, nil
}

func (database *DB) Close() error {
	if err := database.Existence.Close(); err != nil {
		return fmt.Errorf("Close Existence DB: %w", err)
	}

	if err := database.Found.Close(); err != nil {
		return fmt.Errorf("Close Found DB: %w", err)
	}

	if err := database.Ban.Close(); err != nil {
		return fmt.Errorf("Close Db DB: %w", err)
	}

	database.Existence = nil
	database.Found = nil
	database.Ban = nil

	return nil
}

// Ban a url for a specific reason.
func (database *DB) ban(u *url.URL, reason string) {
	database.Ban.Add(u)
	database.save(u, &Page{Error: reason})
}

// Save a page in the (existence and object) DB.
func (database *DB) save(u *url.URL, page *Page) {
	key := db.NewURLKey(u)

	page.Time = time.Now().UTC()
	page.URL = *u

	database.KeyValueDB.Store(key, page)
}
