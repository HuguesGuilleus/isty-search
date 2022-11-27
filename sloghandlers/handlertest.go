package sloghandlers

import (
	"bytes"
	"fmt"
	"golang.org/x/exp/slog"
)

type handlerTest struct {
	// All Records.
	// Recod.Time and Context is set to zero.
	records *[]string

	// Minimal level that store records.
	level slog.Level

	// intern attributes to to the future records.
	attributeSlice []slog.Attr
	// The name of the
	attributeGroup string
}

// A slog.Handler to store the records in a slice of string.
// Use it to test a package.
type HandlerRecords interface {
	slog.Handler
	// Each record have a line of the format:
	//	LEVEL [message] group.subgroup.key=value ...
	Records() []string
}

func NewHandlerRecords(level slog.Level) HandlerRecords {
	return &handlerTest{
		level:   level,
		records: new([]string),
	}
}

func (h *handlerTest) Records() []string { return *h.records }

func (h *handlerTest) Enabled(l slog.Level) bool { return h.level <= l }

func (h *handlerTest) Handle(r slog.Record) error {
	if !h.Enabled(r.Level) {
		return nil
	}

	buff := bytes.Buffer{}
	fmt.Fprintf(&buff, "%s [%s]", r.Level, r.Message)

	printAttr := func(string, slog.Attr) {}
	printAttr = func(base string, a slog.Attr) {
		if a.Value.Kind() == slog.GroupKind {
			base += a.Key + "."
			for _, subAttr := range a.Value.Group() {
				printAttr(base, subAttr)
			}
		} else {
			buff.WriteByte(' ')
			buff.WriteString(base)
			buff.WriteString(a.String())
		}
	}
	for _, a := range h.attributeSlice {
		printAttr("", a)
	}
	r.Attrs(func(a slog.Attr) { printAttr(h.attributeGroup, a) })

	*h.records = append(*h.records, buff.String())

	return nil
}

func (h *handlerTest) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	if h.attributeGroup == "" {
		newHandler.attributeSlice = append(newHandler.attributeSlice, attrs...)
	} else {
		newHandler.attributeSlice = append(newHandler.attributeSlice, slog.Group(h.attributeGroup[:len(h.attributeGroup)-1], attrs...))
	}
	return &newHandler
}

func (h *handlerTest) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.attributeGroup += name + "."
	return &newHandler
}
