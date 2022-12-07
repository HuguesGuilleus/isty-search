package sloghandlers

import (
	"bytes"
	"fmt"
	"golang.org/x/exp/slog"
	"io"
	"os"
	"strings"
	"sync"
)

const barLength = 50

func NewConsole(level slog.Level) slog.Handler {
	return &commonHandler{
		level: level,
		wrap:  &console{outWriter: os.Stderr},
	}
}

type console struct {
	outWriter io.Writer
	outMutex  sync.Mutex

	bar string
}

func (c *console) Begin(buff *bytes.Buffer, r slog.Record) bool {
	if c.bar != "" {
		buff.WriteString("\033[1G\033[K")
	}

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

		percentage := 0
		if l > 0 {
			percentage = i * 100 / l
		}
		fmt.Fprintf(buff, " %3d %% ", percentage)

		rate := barLength
		if l > 0 {
			rate = i * barLength / l
		}
		for i := 0; i < rate; i++ {
			buff.WriteRune('█')
		}
		for i := rate; i < barLength; i++ {
			buff.WriteRune('░')
		}

		c.bar = strings.Clone(buff.String())

		c.outMutex.Lock()
		defer c.outMutex.Unlock()
		buff.WriteTo(c.outWriter)

		return true

	case "%end":
		c.bar = ""

		c.outMutex.Lock()
		defer c.outMutex.Unlock()
		c.outWriter.Write([]byte("\033[1G\033[K"))

		return true

	default:
		fmt.Fprintf(buff, "%s [%s]", r.Level, r.Message)
		return false
	}
}

func (c *console) Store(buff *bytes.Buffer) error {
	buff.WriteByte('\n')
	if c.bar != "" {
		buff.WriteString(c.bar)
	}

	c.outMutex.Lock()
	defer c.outMutex.Unlock()

	_, err := buff.WriteTo(c.outWriter)
	return err
}
