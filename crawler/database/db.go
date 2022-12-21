package crawldatabase

import (
	"errors"
	"fmt"
	"golang.org/x/exp/slog"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type Database[T any] interface {
	// Add unknwon url.
	// If the URL is known, is deleted of urls, else is saved is DB files.
	// Error are logged and returned.
	AddURL(map[Key]*url.URL) error

	// Close the database.
	// After close, call of database method can infinity block.
	Close() error
}

var NotExist = errors.New("Not exist")

const (
	filenameURLS = "urls.txt"
	filenameMeta = "urls.meta"
	filenameData = "file-0.gz"
)

type database[T any] struct {
	logger *slog.Logger

	// The base path.
	base string

	mutex    sync.Mutex
	mapMeta  map[Key]metavalue
	metaFile *os.File
	urlsFile *os.File
	dataFile *os.File
}

// Open the DB, and return all know URL.
func OpenWithKnow[T any](logger *slog.Logger, base string) ([]*url.URL, Database[T], error) {
	return open[T](logger, base, []byte{TypeKnow})
}

func open[T any](logger *slog.Logger, base string, acceptedTypes []byte) ([]*url.URL, Database[T], error) {
	base = filepath.Clean(base)
	if err := os.MkdirAll(base, 0o775); err != nil {
		logger.Error("db.open", err, "mkdir", base)
		return nil, nil, err
	}

	mapMeta := loadElasticMetavalue(readFile(logger, base, filenameMeta))
	urls := loadURLs(logger, readFile(logger, base, filenameURLS), mapMeta, acceptedTypes)

	metaFile, err := openFile(logger, base, filenameMeta, os.O_WRONLY)
	if err != nil {
		return nil, nil, err
	}
	urlsFile, err := openFile(logger, base, filenameURLS, os.O_WRONLY)
	if err != nil {
		return nil, nil, err
	}
	dataFile, err := openFile(logger, base, filenameData, os.O_RDWR)
	if err != nil {
		return nil, nil, err
	}

	logger.Info("db.open", "base", base)
	getStatistics(mapMeta).Log(logger)

	return urls, &database[T]{
		logger:   logger,
		base:     base,
		mapMeta:  mapMeta,
		metaFile: metaFile,
		urlsFile: urlsFile,
		dataFile: dataFile,
	}, nil
}

// Open the file "base/name" and log error if occure.
func openFile(logger *slog.Logger, base, name string, flag int) (f *os.File, err error) {
	path := filepath.Join(base, name)
	f, err = os.OpenFile(path, flag|os.O_APPEND|os.O_CREATE, 0o664)

	if err != nil {
		logger.Error("db.open", err, "file", path)
	}

	return
}

// Read all file "base/name" and log error if occure.
func readFile(logger *slog.Logger, base, name string) []byte {
	path := filepath.Join(base, name)
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		logger.Error("db.readfile", err, "file", path)
	}
	return data
}

func (db *database[_]) Close() error {
	db.mutex.Lock() // Keep locked to block the database

	errs := []error{
		db.metaFile.Close(),
		db.urlsFile.Close(),
		db.dataFile.Close(),
	}
	finalErr := error(nil)
	for _, err := range errs {
		if err != nil {
			db.logger.Error("db.close", err)
			finalErr = err
		}
	}

	return finalErr
}

// Return statistics of the database
func (db *database[_]) Statistics() Statistics {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return getStatistics(db.mapMeta)
}

func (db *database[_]) AddURL(urls map[Key]*url.URL) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	for key, u := range urls {
		if db.mapMeta[key].Type == TypeNothing {
			if _, err := db.urlsFile.WriteString(u.String() + "\n"); err != nil {
				f := filepath.Join(db.base, filenameURLS)
				db.logger.Error("db.err", err, "file", f)
				return fmt.Errorf("DB Write in %q: %w", f, err)
			}

			meta := metavalue{Type: TypeKnow}
			if err := writeElasticMetavalue(key, meta, db.metaFile); err != nil {
				f := filepath.Join(db.base, filenameMeta)
				db.logger.Error("db.err", err, "file", f)
				return fmt.Errorf("DB Write in %q: %w", f, err)
			}
			db.mapMeta[key] = meta
		} else {
			delete(urls, key)
		}
	}

	return nil
}
