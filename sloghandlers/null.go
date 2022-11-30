package sloghandlers

import (
	"golang.org/x/exp/slog"
)

type nullHandler struct{}

func NewNullHandler() slog.Handler                       { return nullHandler{} }
func (_ nullHandler) Enabled(slog.Level) bool            { return false }
func (_ nullHandler) Handle(slog.Record) error           { return nil }
func (_ nullHandler) WithAttrs([]slog.Attr) slog.Handler { return nullHandler{} }
func (_ nullHandler) WithGroup(string) slog.Handler      { return nullHandler{} }
