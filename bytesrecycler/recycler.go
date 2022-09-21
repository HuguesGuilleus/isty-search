package recycler

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func Recycle(buffer *bytes.Buffer) {
	buffer.Reset()
	pool.Put(buffer)
}
