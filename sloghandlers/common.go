package sloghandlers

import (
	"bytes"
	"golang.org/x/exp/slog"
	"strconv"
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

	for _, a := range h.attributeSlice {
		writeAttr(&buff, "", a)
	}
	r.Attrs(func(a slog.Attr) { writeAttr(&buff, h.attributeGroup, a) })

	return h.wrap.Store(&buff)
}

func writeAttr(buff *bytes.Buffer, base string, a slog.Attr) {
	switch a.Value.Kind() {
	case slog.GroupKind:
		base += a.Key + "."
		for _, subAttr := range a.Value.Group() {
			writeAttr(buff, base, subAttr)
		}
	case slog.Int64Kind:
		writeKey(buff, base, a.Key)
		i := a.Value.Int64()
		if i < 0 {
			buff.WriteByte('-')
			printUint64(buff, uint64(-i))
		} else {
			buff.WriteByte('+')
			printUint64(buff, uint64(i))
		}
	case slog.Uint64Kind:
		writeKey(buff, base, a.Key)
		printUint64(buff, a.Value.Uint64())
	default:
		writeKey(buff, base, a.Key)
		buff.WriteString(a.Value.String())
	}
}

func writeKey(buff *bytes.Buffer, base, key string) {
	buff.WriteByte(' ')
	buff.WriteString(base)
	buff.WriteString(key)
	buff.WriteByte('=')
}

func printUint64(buff *bytes.Buffer, u uint64) {
	if v := u / 1000; v > 0 {
		printUint64(buff, v)
		buff.WriteByte('_')
	}
	s := strconv.FormatUint(u%1000, 10)
	for i := len(s); i < 3; i++ {
		buff.WriteByte('0')
	}
	buff.WriteString(s)
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
