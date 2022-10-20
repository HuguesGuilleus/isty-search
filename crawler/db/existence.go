package db

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
)

type Existence struct {
	file  *os.File
	keys  map[Key]bool
	mutex sync.RWMutex
}

func OpenExistence(filePath string) (*Existence, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o665)
	if err != nil {
		return nil, fmt.Errorf("OpenExistence(%q): %w", filePath, err)
	}

	db := &Existence{
		keys: make(map[Key]bool),
		file: f,
	}

	if err := readKeys(f, func(key Key) { db.keys[key] = true }); err != nil {
		return nil, fmt.Errorf("OpenExistence(%q): %w", filePath, err)
	}

	return db, nil
}

// Put a key into the Existence db
func (db *Existence) Add(key Key) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	if _, err := db.file.Write(key[:]); err != nil {
		return fmt.Errorf("Existence.Add(): %w", err)
	}

	db.keys[key] = true

	return nil
}

// Check if a key exist in the db
func (db *Existence) Exist(key Key) bool {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.keys[key]
}

// Filter remove known url from the list, and remove duplicated URL.
func (db *Existence) Filter(list []*url.URL) []*url.URL {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	m := make(map[string]*url.URL, len(list))
	for _, u := range list {
		if db.keys[NewURLKey(u)] {
			continue
		}
		m[u.String()] = u
	}

	filtered := make([]*url.URL, 0, len(list))
	for _, u := range m {
		filtered = append(filtered, u)
	}

	return filtered
}

// Close the internal file and the map of keys. Must no use the DB after
// a call to this method, because all future call will be blocked.
func (db *Existence) Close() error {
	db.mutex.Lock()
	// Keep block the db.
	db.keys = nil
	return db.file.Close()
}

func ReadExistence(filePath string, numberOfKeys *int, fn func(key Key)) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ReadExistence(%q) %w", filePath, err)
	}
	defer file.Close()

	if numberOfKeys != nil {
		info, err := file.Stat()
		if err != nil {
			return fmt.Errorf("ReadExistence(%q) %w", filePath, err)
		}
		*numberOfKeys = int(info.Size() / sha256.Size)
	}

	if err := readKeys(file, fn); err != nil {
		return fmt.Errorf("ReadExistence(%q) %w", filePath, err)
	}

	return nil
}

// Read all key from the file, for each call fn.
func readKeys(file *os.File, fn func(key Key)) error {
	for {
		key := Key{}
		if _, err := file.Read(key[:]); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		fn(key)
	}
}
