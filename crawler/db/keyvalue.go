package db

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/bytesrecycler"
	"io/fs"
	"os"
	"path/filepath"
)

type KeyValueDB[T any] struct {
	base string
}

func OpenKeyValueDB[T any](base string) KeyValueDB[T] {
	base = filepath.Clean(base)
	os.MkdirAll(base, 0o775)
	return KeyValueDB[T]{base: base}
}

func (db *KeyValueDB[T]) Store(key Key, v *T) error {
	buffGob := recycler.Get()
	defer recycler.Recycle(buffGob)

	if err := gob.NewEncoder(buffGob).Encode(v); err != nil {
		return fmt.Errorf("KeyValueDB.Store(%x): Gob encode: %w", key, err)
	}

	buffGzip := recycler.Get()
	defer recycler.Recycle(buffGzip)

	gzipWriter := gzip.NewWriter(buffGzip)
	if _, err := buffGob.WriteTo(gzipWriter); err != nil {
		return fmt.Errorf("KeyValueDB.Store(%x): Gzip: %w", key, err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("KeyValueDB.Store(%x): Gzip: %w", key, err)
	}

	os.MkdirAll(key.dir(db.base), 0o775)

	if err := os.WriteFile(key.path(db.base), buffGzip.Bytes(), 0o664); err != nil {
		return fmt.Errorf("KeyValueDB.Store(%x): os write file: %w", key, err)
	}

	return nil
}

func (db *KeyValueDB[T]) Get(key Key) (*T, error) {
	data, err := os.ReadFile(key.path(db.base))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("KeyValueDB.Get(%s): Not found", key)
		}
		return nil, fmt.Errorf("KeyValueDB.Get(%s): %w", key, err)
	}

	buff := recycler.Get()
	defer recycler.Recycle(buff)

	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("KeyValueDB.Get(%x), uncompress with gzip: %w", key, err)
	}
	if _, err := buff.ReadFrom(gzipReader); err != nil {
		return nil, fmt.Errorf("KeyValueDB.Get(%x), uncompress with gzip: %w", key, err)
	}
	if err = gzipReader.Close(); err != nil {
		return nil, fmt.Errorf("KeyValueDB.Get(%x), uncompress with gzip: %w", key, err)
	}

	v := new(T)
	if err := gob.NewDecoder(buff).Decode(v); err != nil {
		return nil, fmt.Errorf("KeyValueDB.Get(%x), gob decode: %w", key, err)
	}

	return v, nil
}
