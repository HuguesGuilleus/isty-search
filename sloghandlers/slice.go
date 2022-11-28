package sloghandlers

import (
	"golang.org/x/exp/slog"
)

type multiHandler []slog.Handler

// Create a handlers that combine multiple handler.
func NewMultiHandlers(handlers ...slog.Handler) slog.Handler { return multiHandler(handlers) }

func (m multiHandler) Enabled(l slog.Level) bool {
	if len(m) > 0 {
		return m[0].Enabled(l)
	}
	return false
}

func (m multiHandler) Handle(r slog.Record) error {
	for _, h := range m {
		if err := h.Handle(r); err != nil {
			return err
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := make(multiHandler, len(m))
	for i, h := range m {
		newHandler[i] = h.WithAttrs(attrs)
	}
	return newHandler
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	newHandler := make(multiHandler, len(m))
	for i, h := range m {
		newHandler[i] = h.WithGroup(name)
	}
	return newHandler
}
