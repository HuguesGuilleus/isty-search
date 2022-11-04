package db

import (
	"fmt"
	"io"
	"net/url"
	"os"
)

type Existence interface {
	// Put a key.
	Add(key Key) error
	// Check if a key exists.
	Exist(key Key) bool
	// Filter remove known url from the list, and remove duplicated URL.
	Filter(list []*url.URL) []*url.URL
	// Close the DB (internal file, remove map of keys...).
	// Must no use the DB after a call to this method,
	// because all future call will be blocked.
	Close() error
}

type existenceFile struct {
	file *os.File
	existenceMap
}

// Open a Existence DB from a file.
// Key are stored into memory to speed access and append to the file.
// The file is a succession of key without any order.
func OpenExistenceFile(filePath string) (Existence, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o665)
	if err != nil {
		return nil, fmt.Errorf("OpenExistenceFile(%q): %w", filePath, err)
	}

	keys, err := loadKeys(f)
	if err != nil {
		return nil, fmt.Errorf("OpenExistenceFile(%q): %w", filePath, err)
	}

	return &existenceFile{
		file:         f,
		existenceMap: existenceMap{keys: keys},
	}, nil
}

func (db *existenceFile) Add(key Key) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if db.keys[key] {
		return nil
	}
	db.keys[key] = true

	if _, err := db.file.Write(key[:]); err != nil {
		return fmt.Errorf("Existence.Add(): %w", err)
	}

	return nil
}

func (db *existenceFile) Close() error {
	db.existenceMap.Close()
	return db.file.Close()
}

// Load existence file, and return a map with all key.
// Map value is always true.
func LoadExistenceFile(filePath string) (map[Key]bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("LoadExistenceFile(%q): %w", filePath, err)
	}
	defer file.Close()

	keys, err := loadKeys(file)
	if err != nil {
		return nil, fmt.Errorf("LoadExistenceFile(%q): %w", filePath, err)
	}
	return keys, nil
}

// Read all key from the file, and store it into the map.
func loadKeys(file *os.File) (map[Key]bool, error) {
	m := make(map[Key]bool)
	for {
		key := Key{}
		if _, err := file.Read(key[:]); err == io.EOF {
			return m, nil
		} else if err != nil {
			return nil, err
		}
		m[key] = true
	}
}
