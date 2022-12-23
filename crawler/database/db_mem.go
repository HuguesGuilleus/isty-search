package crawldatabase

import (
	"golang.org/x/exp/slog"
	"io"
	"time"
)

// Open a database in the memory, so it not persistent.
// Use only for test.
func OpenMem[T any](logger *slog.Logger) Database[T] {
	return &database[T]{
		logger:      logger,
		statsTicker: &time.Ticker{},
		base:        "$memory",
		mapMeta:     make(map[Key]metavalue),
		metaFile:    &memFile{},
		urlsFile:    &memFile{},
		dataFile:    &memFile{},
		position:    0,
	}
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
