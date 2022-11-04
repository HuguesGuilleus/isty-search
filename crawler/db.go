package crawler

import (
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler/db"
	"path/filepath"
)

// A composite DB with sub subdatabase for specific usage.
type DB struct {
	Object    db.ObjectBD[Page]
	Existence db.Existence
	URLs      *db.URLsDB
	Ban       *db.BanURL
}

func OpenDB(root string) (*DB, error) {
	existence, err := db.OpenExistenceFile(filepath.Join(root, "existence-key.bin"))
	if err != nil {
		return nil, err
	}

	urls, err := db.OpenURLsDB(filepath.Join(root, "url"))
	if err != nil {
		return nil, err
	}

	ban, err := db.OpenBanURL(filepath.Join(root, "ban.txt"))
	if err != nil {
		return nil, err
	}

	return &DB{
		Object:    db.OpenObjectBD[Page](filepath.Join(root, "object")),
		Existence: existence,
		URLs:      urls,
		Ban:       ban,
	}, nil
}

func (database *DB) Close() error {
	if err := database.Existence.Close(); err != nil {
		return fmt.Errorf("Close exitence DB: %w", err)
	}

	if err := database.URLs.Close(); err != nil {
		return fmt.Errorf("Close URLs DB: %w", err)
	}

	if err := database.Ban.Close(); err != nil {
		return fmt.Errorf("Close URLs DB: %w", err)
	}

	database.Existence = nil
	database.URLs = nil
	database.Ban = nil

	return nil
}
