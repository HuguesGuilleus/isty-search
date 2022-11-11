package db

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var IgnorePaths = []string{
	"/robots.txt",
	"/favicon.ico",
}

const (
	valueLen   = KeyLen + 8
	existValue = 1
)

type URLsDB struct {
	maxURLLen int
	mutex     sync.Mutex
	keysMap   map[Key]int64
	keysFile  *os.File
	urlsFile  *os.File
	end       chan<- struct{}
	// Statistics
	nbBan   int
	nbFound int
	nbDone  int
}

func OpenURLsDB(dir string, maxURLLen int) (*URLsDB, []*url.URL, error) {
	dir = filepath.Clean(dir)

	// Load Keys
	keysPath := filepath.Join(dir, "urls.key")
	data, err := os.ReadFile(keysPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, fmt.Errorf("OpenURLsDB: From file (%q) %w", keysPath, err)
	}
	keysMap := urlsDBLoadData(data)
	keysFile, err := os.OpenFile(keysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o664)
	if err != nil {
		return nil, nil, fmt.Errorf("OpenURLsDB: Read(%q) %w", keysPath, err)
	}

	// Load plain URLs
	urlsPath := filepath.Join(dir, "urls.txt")
	data, err = os.ReadFile(urlsPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, nil, fmt.Errorf("OpenURLsDB: Read(%q) %w", urlsPath, err)
	}
	urls, err := loadURLS(data, keysMap)
	if err != nil {
		return nil, nil, fmt.Errorf("OpenURLsDB: From file %q: %w", urlsPath, err)
	}
	urlsFile, err := os.OpenFile(urlsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o664)
	if err != nil {
		return nil, nil, fmt.Errorf("OpenURLsDB: Read(%q) %w", urlsPath, err)
	}

	end := make(chan struct{})
	db := &URLsDB{
		maxURLLen: maxURLLen,
		keysMap:   keysMap,
		keysFile:  keysFile,
		urlsFile:  urlsFile,
		end:       end,
	}

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				db.stats()
				fmt.Printf("[DB STATS] done:%d, ban:%d, found:%d, total:%d\n", db.nbDone, db.nbBan, db.nbFound, len(db.keysMap))
			case <-end:
				return
			}
		}
	}()

	return db, urls, nil
}

func loadURLS(data []byte, keysMap map[Key]int64) ([]*url.URL, error) {
	urls := make([]*url.URL, 0)
	for i, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		switch keysMap[NewStringKey(line)] {
		case 0, existValue:
		default:
			continue
		}
		if !utf8.ValidString(line) {
			return nil, fmt.Errorf("Line %d, %q: Invalid utf8 string", i, line)
		}
		u, err := url.Parse(line)
		if err != nil {
			return nil, fmt.Errorf("Parse URL line %d, %q: %w", i, line, err)
		}
		urls = append(urls, u)
	}

	return urls, nil
}

func (db *URLsDB) Close() error {
	db.mutex.Lock() // Leave blocked

	db.end <- struct{}{}
	close(db.end)
	db.end = nil

	defer func() {
		db.keysMap = nil
		db.keysFile = nil
		db.urlsFile = nil
	}()

	const format = "Close URLsDB: %w"
	if err := db.keysFile.Close(); err != nil {
		return fmt.Errorf(format, err)
	}
	if err := db.urlsFile.Close(); err != nil {
		return fmt.Errorf(format, err)
	}

	return nil
}

func (db *URLsDB) Merge(urls []*url.URL) ([]*url.URL, error) {
	// Unique url, and remove ignred path
	urlsMap := make(map[string]*url.URL, len(urls))
forURLs:
	for _, u := range urls {
		for _, u := range getParentURL(u) {
			for _, ignorePath := range IgnorePaths {
				if u.Path == ignorePath {
					continue forURLs
				}
			}
			s := u.String()
			if len(s) >= db.maxURLLen {
				continue
			}
			urlsMap[s] = u
		}
	}

	// Skip Mutex Lock, and returned allocation
	if len(urlsMap) == 0 {
		return nil, nil
	}

	// Store URL
	db.mutex.Lock()
	defer db.mutex.Unlock()
	returned := make([]*url.URL, 0, len(urlsMap))
	for s, u := range urlsMap {
		key := NewStringKey(s)
		if db.keysMap[key] != 0 {
			continue
		}
		returned = append(returned, u)

		if _, err := db.urlsFile.WriteString(s + "\n"); err != nil {
			return nil, fmt.Errorf("URLsDB store %q: %w", s, err)
		}

		if err := db.store(key, existValue, false); err != nil {
			return nil, err
		}
	}

	return returned, nil
}

func getParentURL(src *url.URL) []*url.URL {
	u := *src
	nbCut := strings.Count(u.Host, ".") - 1
	urls := make([]*url.URL, 1, nbCut+3)

	urls[0] = src
	if u.Path != "/" {
		u.Path = "/"
		newURL := u
		urls = append(urls, &newURL)
	}

	if newHost, _, cutted := strings.Cut(u.Host, ":"); cutted {
		u.Host = newHost
		newURL := u
		urls = append(urls, &newURL)
	}

	for i := 0; i < nbCut; i++ {
		u.Host = u.Host[strings.IndexByte(u.Host, '.')+1:]
		newURL := u
		urls = append(urls, &newURL)
	}

	return urls
}

func (db *URLsDB) Store(key Key, banned bool) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.store(key, time.Now().Unix(), banned)
}

func (db *URLsDB) store(key Key, value int64, banned bool) error {
	if banned {
		value = -value
		db.nbBan++
	} else if value == existValue {
		db.nbFound++
	} else {
		db.nbDone++
	}
	db.keysMap[key] = value

	data := [valueLen]byte{}
	copy(data[:], key[:])
	data[KeyLen+0] = byte(value >> 0o70 & 0xFF)
	data[KeyLen+1] = byte(value >> 0o60 & 0xFF)
	data[KeyLen+2] = byte(value >> 0o50 & 0xFF)
	data[KeyLen+3] = byte(value >> 0o40 & 0xFF)
	data[KeyLen+4] = byte(value >> 0o30 & 0xFF)
	data[KeyLen+5] = byte(value >> 0o20 & 0xFF)
	data[KeyLen+6] = byte(value >> 0o10 & 0xFF)
	data[KeyLen+7] = byte(value >> 0o00 & 0xFF)

	_, err := db.keysFile.Write(data[:])
	if err != nil {
		return fmt.Errorf("URLsDB store %x key: %w", key, err)
	}

	return nil
}

func (db *URLsDB) stats() {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	db.nbDone = 0
	db.nbFound = 0
	db.nbBan = 0
	for _, value := range db.keysMap {
		switch {
		case value == 0:
			// ignore
		case value < 0:
			db.nbBan++
		case value == existValue:
			db.nbFound++
		default:
			db.nbDone++
		}
	}
}

func urlsDBLoadData(data []byte) map[Key]int64 {
	keysMap := make(map[Key]int64, len(data)/valueLen)
	for i := 0; i+valueLen <= len(data); i += valueLen {
		key := Key{}
		copy(key[:], data[i:])
		keysMap[key] = 0 |
			int64(data[i+KeyLen+0])<<0o70 |
			int64(data[i+KeyLen+1])<<0o60 |
			int64(data[i+KeyLen+2])<<0o50 |
			int64(data[i+KeyLen+3])<<0o40 |
			int64(data[i+KeyLen+4])<<0o30 |
			int64(data[i+KeyLen+5])<<0o20 |
			int64(data[i+KeyLen+6])<<0o10 |
			int64(data[i+KeyLen+7])<<0o00
	}
	return keysMap
}

// Call f on each done key.
func (db *URLsDB) ForDone(f func(Key)) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	for key, value := range db.keysMap {
		if value > existValue {
			f(key)
		}
	}
}
