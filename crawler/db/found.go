package db

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"sync"
)

type Found struct {
	mutex  sync.Mutex
	founed map[Key]bool
	file   *os.File
}

func OpenFound(filePath string) (*Found, []*url.URL, error) {
	founed := make(map[Key]bool)
	urls := make([]*url.URL, 0)

	err := ReadBanURL(filePath, func(u *url.URL) {
		founed[NewURLKey(u)] = true
		urls = append(urls, u)
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, fmt.Errorf("OpenFoundDB(%q): %w", filePath, err)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o664)
	if err != nil {
		return nil, nil, fmt.Errorf("OpenFoundDB(%q) %w", filePath, err)
	}

	return &Found{
		founed: founed,
		file:   f,
	}, urls, nil
}

func (db *Found) AddURLs(urls []*url.URL) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	for _, u := range urls {
		key := NewURLKey(u)
		if db.founed[key] {
			continue
		}
		db.founed[key] = true
		db.file.WriteString(u.String() + "\n")
	}
}

func (db *Found) Close() error {
	db.mutex.Lock() // Block the DB
	db.founed = nil
	return db.file.Close()
}
