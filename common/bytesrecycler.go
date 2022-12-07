package common

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func GetBuffer() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func RecycleBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	pool.Put(buffer)
}
