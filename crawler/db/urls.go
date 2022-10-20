package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxURLinFile = 4096 // arbitrary

// Data Base to store URL that will be processed.
type URLsDB struct {
	root string

	currentMutex sync.Mutex
	currentPath  string
	currentFile  *os.File
	currentLen   int
	currentURLs  []*url.URL

	previousMutex sync.Mutex
	previousFiles []string
}

func OpenURLsDB(root string) (*URLsDB, error) {
	db := &URLsDB{
		root: filepath.Clean(root),
	}
	db.currentPath = URLsDBgetPath(db.root)
	err := error(nil)
	db.currentFile, err = os.Create(db.currentPath)
	if err != nil {
		return nil, fmt.Errorf("OpenURLsDB(%q): %w", db.root, err)
	}
	db.currentURLs = make([]*url.URL, 0, maxURLinFile)

	previsousEntry, err := os.ReadDir(db.root)
	if err != nil {
		return nil, fmt.Errorf("OpenURLsDB(%q): %w", db.root, err)
	}
	for i, entry := range previsousEntry {
		db.previousFiles[i] = filepath.Join(db.root, entry.Name())
	}

	return db, nil
}

// Add some URL into the db.
func (db *URLsDB) Add(urls []*url.URL) {
	db.currentMutex.Lock()
	defer db.currentMutex.Unlock()

	for _, u := range urls {
		db.currentFile.WriteString(u.String() + "\n")
		db.currentURLs = append(db.currentURLs, u)
		db.currentLen++

		if db.currentLen >= maxURLinFile {
			db.currentFile.Close()
			func(path string) {
				db.previousMutex.Lock()
				defer db.previousMutex.Unlock()
				db.previousFiles = append(db.previousFiles, path)
			}(db.currentPath)

			db.currentPath = URLsDBgetPath(db.root)
			db.currentFile, _ = os.Create(db.currentPath)
			db.currentLen = 0
			db.currentURLs = db.currentURLs[:0]
		}
	}
}

func (db *URLsDB) Close() error {
	db.currentMutex.Lock()
	db.previousMutex.Lock()
	// do not unlock to block the DB

	return db.currentFile.Close()
}

func (db *URLsDB) Get() []*url.URL {
	db.currentMutex.Lock()
	defer db.currentMutex.Unlock()

	if len(db.currentURLs) > 0 {
		urls := make([]*url.URL, len(db.currentURLs))
		copy(urls, db.currentURLs)
		db.currentURLs = db.currentURLs[:0]
		return urls
	}

	db.previousMutex.Lock()
	defer db.previousMutex.Unlock()

	if len(db.previousFiles) == 0 {
		return nil
	}

	urls := loadURLs(db.previousFiles[0])
	db.previousFiles = db.previousFiles[1:]
	return urls
}

func loadURLs(filePath string) []*url.URL {
	content, _ := os.ReadFile(filePath)
	urls := make([]*url.URL, 0, maxURLinFile)
	for _, line := range strings.Split(string(content), "\n") {
		u, _ := url.Parse(line)
		if u == nil {
			continue
		}
		urls = append(urls, u)
	}
	return urls
}

// Get the path to store url.
func URLsDBgetPath(root string) string {
	return fmt.Sprintf("%s%c%d.txt", root, filepath.Separator, time.Now().Unix())
}
