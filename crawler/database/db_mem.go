package crawldatabase

import (
	"github.com/HuguesGuilleus/isty-search/sloghandlers"
	"golang.org/x/exp/slog"
	"io"
	"net/url"
	"time"
)

// Open a database in the memory, so it not persistent.
// Use only for test.
// Always retuns nil for url slice and error.
func OpenMemory[T any](logger *slog.Logger, _ string, _ bool) ([]*url.URL, *Database[T], error) {
	if logger == nil {
		logger = slog.New(sloghandlers.NewNullHandler())
	}

	return nil, &Database[T]{
		logger:      logger,
		statsTicker: &time.Ticker{},
		base:        "$memory",
		mapMeta:     make(map[Key]metavalue),
		metaFile:    &memFile{},
		urlsFile:    &memFile{},
		dataFile:    &memFile{},
		position:    0,
	}, nil
}

type memFile []byte

func (f *memFile) Close() error {
	f = nil
	return nil
}
func (f *memFile) ReadAt(p []byte, off int64) (n int, err error) {
	if int(off)+len(p) > len(*f) {
		return 0, io.ErrShortBuffer
	}
	return copy(p, (*f)[int(off):]), nil
}
func (f *memFile) WriteString(s string) (int, error) { return f.Write([]byte(s)) }
func (f *memFile) Write(p []byte) (int, error) {
	*f = append(*f, p...)
	return len(p), nil
}
