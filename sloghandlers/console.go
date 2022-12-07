package sloghandlers

import (
	"bytes"
	"fmt"
	"golang.org/x/exp/slog"
	"io"
	"os"
	"sync"
)

const (
	barClean  = "\033[1G\033[K"
	barLength = 50
)

// Create a slog.Handler that print log element into the console.
//
// Manage console bar:
//   - "%" and attributes: "%i" and "%len" (optinal)
//   - "%end", attribute are ignored.
func NewConsole(level slog.Level) slog.Handler {
	return &commonHandler{
		level: level,
		wrap:  &console{outWriter: os.Stderr},
	}
}

type console struct {
	outWriter io.Writer
	outMutex  sync.Mutex

	bar bytes.Buffer
}

func (c *console) Begin(buff *bytes.Buffer, r slog.Record) bool {
	switch r.Message {
	case "%":
		i := 0
		l := 0
		r.Attrs(func(a slog.Attr) {
			switch a.Key {
			case "%i":
				i = int(a.Value.Int64())
			case "%len":
				l = int(a.Value.Int64())
			}
		})

		c.bar.Reset()
		c.bar.WriteString(barClean)

		percentage := 0
		if l > 0 {
			percentage = i * 100 / l
		}
		fmt.Fprintf(&c.bar, " %3d %% ", percentage)

		rate := barLength
		if l > 0 {
			rate = i * barLength / l
		}
		for i := 0; i < rate; i++ {
			c.bar.WriteRune('█')
		}
		for i := rate; i < barLength; i++ {
			c.bar.WriteRune('░')
		}

		c.outMutex.Lock()
		defer c.outMutex.Unlock()
		c.outWriter.Write(c.bar.Bytes())

		return true

	case "%end":
		c.bar.Reset()

		c.outMutex.Lock()
		defer c.outMutex.Unlock()
		c.outWriter.Write([]byte(barClean))

		return true

	default:
		if c.bar.Len() != 0 {
			buff.WriteString(barClean)
		}
		fmt.Fprintf(buff, "%s [%s]", r.Level, r.Message)
		return false
	}
}

func (c *console) Store(buff *bytes.Buffer) error {
	buff.WriteByte('\n')
	if c.bar.Len() != 0 {
		buff.Write(c.bar.Bytes())
	}

	c.outMutex.Lock()
	defer c.outMutex.Unlock()

	_, err := buff.WriteTo(c.outWriter)
	return err
}
