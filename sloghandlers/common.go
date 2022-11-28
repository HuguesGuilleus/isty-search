package sloghandlers

import (
	"bytes"
	"golang.org/x/exp/slog"
)

type commonHandler struct {
	wrap interface {
		// Add name, date... before the attributes.
		// Return true to skip.
		Begin(*bytes.Buffer, slog.Record) bool
		// Store the buffer.
		Store(*bytes.Buffer) error
	}

	// Minimal level that store records.
	level slog.Level

	// intern attributes to to the future records.
	attributeSlice []slog.Attr
	// The name of the
	attributeGroup string
}

func (h *commonHandler) Handle(r slog.Record) error {
	if !h.Enabled(r.Level) {
		return nil
	}

	buff := bytes.Buffer{}
	if h.wrap.Begin(&buff, r) {
		return nil
	}

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

	return h.wrap.Store(&buff)
}

func (h *commonHandler) Enabled(l slog.Level) bool { return h.level <= l }

func (h *commonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	if h.attributeGroup == "" {
		newHandler.attributeSlice = append(newHandler.attributeSlice, attrs...)
	} else {
		newHandler.attributeSlice = append(newHandler.attributeSlice, slog.Group(h.attributeGroup[:len(h.attributeGroup)-1], attrs...))
	}
	return &newHandler
}

func (h *commonHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.attributeGroup += name + "."
	return &newHandler
}
