package db

import (
	"net/url"
	"sync"
)

type existenceMap struct {
	keys  map[Key]bool
	mutex sync.RWMutex
}

// Open existence DB into the memory (so no persistent).
// Use it for test.
func OpenExistenceMap() Existence {
	return &existenceMap{keys: make(map[Key]bool)}
}

func (db *existenceMap) Close() error {
	db.mutex.Lock() // keep blocked
	db.keys = nil
	return nil
}

func (db *existenceMap) Add(key Key) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.keys[key] = true
	return nil
}

func (db *existenceMap) Exist(key Key) bool {
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	return db.keys[key]
}

func (db *existenceMap) Filter(list []*url.URL) []*url.URL {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	m := make(map[string]*url.URL, len(list))
	for _, u := range list {
		s := u.String()
		if db.keys[NewStringKey(s)] {
			continue
		}
		m[s] = u
	}

	filtered := make([]*url.URL, 0, len(list))
	for _, u := range m {
		filtered = append(filtered, u)
	}

	return filtered
}
