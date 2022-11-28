package sloghandlers

import (
	"bytes"
	"fmt"
	"golang.org/x/exp/slog"
)

func NewHandlerRecords(level slog.Level) (*[]string, slog.Handler) {
	records := new([]string)
	return records, &commonHandler{
		level: level,
		wrap:  recorder{records},
	}
}

type recorder struct {
	records *[]string
}

func (_ recorder) Begin(buff *bytes.Buffer, r slog.Record) {
	fmt.Fprintf(buff, "%s [%s]", r.Level, r.Message)
}

func (r recorder) Store(buff *bytes.Buffer) error {
	*r.records = append(*r.records, buff.String())
	return nil
}
