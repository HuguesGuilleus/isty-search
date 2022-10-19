package db

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"sync"
)

type BanURL struct {
	file  *os.File
	mutex sync.Mutex
}

func OpenBanURL(filePath string) (*BanURL, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o664)
	if err != nil {
		return nil, fmt.Errorf("OpenBanURL(%q) %w", filePath, err)
	}

	return &BanURL{file: f}, nil
}

func (db *BanURL) Add(u *url.URL) {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	db.file.WriteString(u.String() + "\n")
}

func (db *BanURL) Close() error {
	db.mutex.Lock()
	// Keep loked the DB
	return db.file.Close()
}

func ReadBanURL(filePath string, fn func(*url.URL)) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("ReadBanURL(%q): %w", filePath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		u, err := url.Parse(scanner.Text())
		if err != nil {
			return fmt.Errorf("Parse url %q fail: %w", scanner.Text(), err)
		}
		fn(u)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ReadBanURL(%q): %w", filePath, err)
	}

	return nil
}
