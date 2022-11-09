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

type ObjectBD[T any] struct {
	base string
}

func OpenObjectBD[T any](base string) ObjectBD[T] {
	base = filepath.Clean(base)
	os.MkdirAll(base, 0o775)
	return ObjectBD[T]{base: base}
}

func (db *ObjectBD[T]) Store(key Key, v *T) error {
	buffGob := recycler.Get()
	defer recycler.Recycle(buffGob)

	if err := gob.NewEncoder(buffGob).Encode(v); err != nil {
		return fmt.Errorf("DB.Store(%x): Gob encode: %w", key, err)
	}

	buffGzip := recycler.Get()
	defer recycler.Recycle(buffGzip)

	gzipWriter := gzip.NewWriter(buffGzip)
	if _, err := buffGob.WriteTo(gzipWriter); err != nil {
		return fmt.Errorf("DB.Store(%x): Gzip: %w", key, err)
	}
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("DB.Store(%x): Gzip: %w", key, err)
	}

	os.MkdirAll(key.dir(db.base), 0o775)

	if err := os.WriteFile(key.path(db.base), buffGzip.Bytes(), 0o664); err != nil {
		return fmt.Errorf("DB.Store(%x): os write file: %w", key, err)
	}

	return nil
}

func (db *ObjectBD[T]) Get(key Key) (*T, error) {
	data, err := os.ReadFile(key.path(db.base))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("DB.Get(%x): not found", key)
		}
		return nil, fmt.Errorf("DB.Get(%x): %w", key, err)
	}

	buff := recycler.Get()
	defer recycler.Recycle(buff)

	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("DB.Get(%x), uncompress with gzip: %w", key, err)
	}
	if _, err := buff.ReadFrom(gzipReader); err != nil {
		return nil, fmt.Errorf("DB.Get(%x), uncompress with gzip: %w", key, err)
	}
	if err = gzipReader.Close(); err != nil {
		return nil, fmt.Errorf("DB.Get(%x), uncompress with gzip: %w", key, err)
	}

	v := new(T)
	if err := gob.NewDecoder(buff).Decode(v); err != nil {
		return nil, fmt.Errorf("DB.Get(%x), gob decode: %w", key, err)
	}

	return v, nil
}
