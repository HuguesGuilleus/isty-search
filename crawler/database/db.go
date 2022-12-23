package crawldatabase

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/common"
	"golang.org/x/exp/slog"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// A database object to store value and matadata (time) about a value.
// All method other than Close can be used concurently.
type Database[T any] interface {
	// Add unknwon url.
	// If the URL is known, is deleted of urls, else is saved is DB files.
	// Error are logged and returned.
	AddURL(map[Key]*url.URL) error

	// Get the value from the DB.
	// If the value if not a file, return NotFile.
	// If the value do not exist, return NotExist.
	GetValue(key Key) (*T, error)

	// Set the value to the DB, overwrite previous value.
	// t must be a type of a regular file.
	SetValue(key Key, value *T, t byte) error

	// Set the redirection.
	SetRedirect(key, destination Key) error

	// Set in the DB a simple type: nothing, known or error.
	// Is t is a file type, it return an error, and do not modify the DB.
	SetSimple(key Key, t byte) error

	// Iterate for each element of type TypeFileHTML.
	// The in the logge rthe progression.
	ForHTML(func(key Key, value *T, progress, total int)) error

	// Return all redictions to valid file.
	// If r1 -> r2 -> r3 -> p, the map contain:
	//   - m[r1] = p
	//   - m[r2] = p
	//   - m[r3] = p
	// The redirection chain is limited to 10.
	Redirections() map[Key]Key

	// Close the database.
	// After close, call of database method can infinity block.
	Close() error
}

var (
	NotExist = errors.New("Not exist")
	NotFile  = errors.New("This value is not a file")
)

const (
	filenameURLS = "urls.txt"
	filenameMeta = "urls.meta"
	filenameData = "file-0.gz"
)

type database[T any] struct {
	logger *slog.Logger

	// A ticker to log DB stats at regular interval.
	statsTicker *time.Ticker

	// The base path.
	base string

	mutex    sync.Mutex
	mapMeta  map[Key]metavalue
	metaFile *os.File
	urlsFile *os.File
	dataFile *os.File

	// The position of write in the dataFile, so at end ogf the file.
	position int64
}

// Open the DB, and return all know URL.
func OpenWithKnow[T any](logger *slog.Logger, base string) ([]*url.URL, Database[T], error) {
	return open[T](logger, base, []byte{TypeKnow})
}

// Open the database but return no url.
func Open[T any](logger *slog.Logger, base string) ([]*url.URL, Database[T], error) {
	return open[T](logger, base, nil)
}

func open[T any](logger *slog.Logger, base string, acceptedTypes []byte) ([]*url.URL, Database[T], error) {
	base = filepath.Clean(base)
	if err := os.MkdirAll(base, 0o775); err != nil {
		logger.Error("db.open", err, "mkdir", base)
		return nil, nil, err
	}

	mapMeta := loadElasticMetavalue(readFile(logger, base, filenameMeta))
	urls := []*url.URL(nil)
	if len(acceptedTypes) > 0 {
		urls = loadURLs(logger, readFile(logger, base, filenameURLS), mapMeta, acceptedTypes)
	}

	metaFile, err := openFile(logger, base, filenameMeta, os.O_WRONLY|os.O_APPEND)
	if err != nil {
		return nil, nil, err
	}
	urlsFile, err := openFile(logger, base, filenameURLS, os.O_WRONLY|os.O_APPEND)
	if err != nil {
		return nil, nil, err
	}
	dataFile, err := openFile(logger, base, filenameData, os.O_RDWR)
	if err != nil {
		return nil, nil, err
	}

	position, err := dataFile.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, nil, fmt.Errorf("Open DB, file %q: %w", filepath.Join(base, filenameData), err)
	}

	logger.Info("db.open", "base", base)
	statsTicker := time.NewTicker(time.Second * 30)

	go func() {
		getStatistics(mapMeta).Log(logger)
		for range statsTicker.C {
			getStatistics(mapMeta).Log(logger)
		}
	}()

	return urls, &database[T]{
		logger:      logger,
		statsTicker: statsTicker,
		base:        base,
		mapMeta:     mapMeta,
		metaFile:    metaFile,
		urlsFile:    urlsFile,
		dataFile:    dataFile,
		position:    position,
	}, nil
}

// Open the file "base/name" and log error if occure.
// It add the flag create.
func openFile(logger *slog.Logger, base, name string, flag int) (f *os.File, err error) {
	path := filepath.Join(base, name)
	f, err = os.OpenFile(path, flag|os.O_CREATE, 0o664)

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
	db.statsTicker.Stop()

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

func (db *database[T]) GetValue(key Key) (*T, error) {
	meta := db.getMetavalue(key)
	if meta.Type == TypeNothing {
		return nil, NotExist
	} else if meta.Type < TypeFile || meta.Type >= TypeError {
		return nil, NotFile
	}
	return db.readValue(key, meta)
}

func (db *database[T]) readValue(key Key, meta metavalue) (*T, error) {
	// Read the data chunck
	data := make([]byte, int(meta.Length))
	_, err := db.dataFile.ReadAt(data, meta.Position)
	if err != nil {
		db.logerror("readChunck", key, err)
		return nil, fmt.Errorf("DB.GetValue(key=%s) %w", key, err)
	}

	// Decompress it
	zlibBuffer := common.GetBuffer()
	defer common.RecycleBuffer(zlibBuffer)
	if zlibReader, err := zlib.NewReader(bytes.NewReader(data)); err != nil {
		db.logerror("zlib.decode.open", key, err)
		return nil, fmt.Errorf("DB.GetValue(key=%s) %w", key, err)
	} else if _, err = zlibBuffer.ReadFrom(zlibReader); err != nil {
		db.logerror("zlib.decode.read", key, err)
		return nil, fmt.Errorf("DB.GetValue(key=%s) %w", key, err)
	} else if err = zlibReader.Close(); err != nil {
		db.logerror("zlib.decode.close", key, err)
		return nil, fmt.Errorf("DB.GetValue(key=%s) %w", key, err)
	}

	// Check hash
	if hash := sha256.Sum256(zlibBuffer.Bytes()); !bytes.Equal(hash[12:], meta.Hash[12:]) {
		db.logerror("chechHash", key, nil)
		return nil, fmt.Errorf("DB.GetValue(key=%s) Wring hash", key)
	}

	// Decode
	value := new(T)
	if err := gob.NewDecoder(zlibBuffer).Decode(value); err != nil {
		db.logerror("gob.decode", key, err)
		return nil, fmt.Errorf("DB.GetValue(key=%s) %w", key, err)
	}

	return value, nil
}

func (db *database[_]) getMetavalue(key Key) metavalue {
	db.mutex.Lock()
	defer db.mutex.Unlock()
	return db.mapMeta[key]
}

func (db *database[T]) SetValue(key Key, value *T, t byte) error {
	if t < TypeFile || t >= TypeError {
		return fmt.Errorf("DB.SetValue(key=%s): The type %d is not for a file", key, t)
	} else if value == nil {
		return fmt.Errorf("DB.SetValue(key=%s): the value is nil", key)
	}

	gobBuffer := common.GetBuffer()
	zlibBuffer := common.GetBuffer()
	defer common.RecycleBuffer(gobBuffer)
	defer common.RecycleBuffer(zlibBuffer)

	if err := gob.NewEncoder(gobBuffer).Encode(value); err != nil {
		db.logerror("encode", key, err)
		return fmt.Errorf("DB.SetValue(key=%s) encode value fail: %w", key, err)
	}
	hash := sha256.Sum256(gobBuffer.Bytes())
	zlibWriter := zlib.NewWriter(zlibBuffer)
	gobBuffer.WriteTo(zlibWriter)
	zlibWriter.Close()

	db.mutex.Lock()
	defer db.mutex.Unlock()

	meta := metavalue{
		Type:     t,
		Time:     time.Now().Unix(),
		Hash:     hash,
		Position: db.position,
		Length:   int32(zlibBuffer.Len()),
	}

	n, err := db.dataFile.Write(zlibBuffer.Bytes())
	if err != nil {
		db.logerror("write.data", key, err)
		return fmt.Errorf("DB.SetValue(key=%s) write data: %w", key, err)
	}
	db.position += int64(n)

	if err := db.setmeta(key, meta); err != nil {
		return fmt.Errorf("DB.SetValue(key=%s) %w", key, err)
	}
	db.mapMeta[key] = meta

	return nil
}

func (db *database[_]) SetSimple(key Key, t byte) error {
	if TypeFile <= t && t < TypeError {
		return fmt.Errorf("Db.SetSimple(key=%s, type=%d) use forbiden type file", key, t)
	}

	db.mutex.Lock()
	defer db.mutex.Unlock()

	if t == TypeNothing {
		if err := db.setmeta(key, metavalue{Type: TypeNothing}); err != nil {
			return fmt.Errorf("db.SetSimple(key=%s) %w", key, err)
		}
		delete(db.mapMeta, key)
	} else {
		meta := metavalue{
			Type: t,
			Time: time.Now().Unix(),
		}
		if err := db.setmeta(key, meta); err != nil {
			return fmt.Errorf("db.SetSimple(key=%s) %w", key, err)
		}
		db.mapMeta[key] = meta
	}

	return nil
}

func (db *database[_]) SetRedirect(key, destination Key) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	meta := metavalue{
		Type: TypeRedirect,
		Time: time.Now().Unix(),
		Hash: destination,
	}
	if err := db.setmeta(key, meta); err != nil {
		return fmt.Errorf("SetRedirect(key=%s) %w", key, err)
	}
	db.mapMeta[key] = meta

	return nil
}

func (db *database[T]) ForHTML(f func(Key, *T, int, int)) error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	items := make([]keymetavalue, 0, len(db.mapMeta))
	for key, meta := range db.mapMeta {
		if meta.Type < TypeFile || TypeError <= meta.Type {
			continue
		}
		items = append(items, keymetavalue{key, meta})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].meta.Position < items[j].meta.Position
	})

	l := len(items)
	for i, item := range items {
		v, err := db.readValue(item.key, item.meta)
		if err != nil {
			return err
		}
		db.logger.Info("%", "%i", i, "%len", l)
		f(item.key, v, i, l)
	}
	db.logger.Info("%end")

	return nil
}

func (db *database[_]) Redirections() map[Key]Key {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	m := make(map[Key]Key)
	for key, meta := range db.mapMeta {
		if meta.Type != TypeRedirect {
			continue
		}
		dest := meta.Hash
		for i := 0; i < 10; i++ {
			newMeta := db.mapMeta[dest]
			t := newMeta.Type
			switch {
			case t == TypeRedirect:
				dest = newMeta.Hash
			case TypeFile <= t && t < TypeError:
				m[key] = dest
				break
			default:
				break
			}
		}
	}

	return m
}

func (db *database[_]) setmeta(key Key, meta metavalue) error {
	if err := writeElasticMetavalue(key, meta, db.metaFile); err != nil {
		db.logerror("write.meta", key, err)
		return fmt.Errorf("write meta: %w", err)
	}
	return nil
}

// The log error.
func (db *database[_]) logerror(op string, key Key, err error) {
	db.logger.Error("db.error", err, "op", op, "key", key.String())
}
