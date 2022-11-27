package sloghandlers

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"io"
	"testing"
)

func TestHandlerTestEnabled(t *testing.T) {
	h := NewHandlerRecords(slog.InfoLevel)

	assert.False(t, h.Enabled(slog.DebugLevel))
	assert.True(t, h.Enabled(slog.InfoLevel))
	assert.True(t, h.Enabled(slog.WarnLevel))

	l := slog.New(h)
	l.Warn("simple", slog.Int("attr", 42))
	l.Error("fatal", io.EOF, slog.Group("group",
		slog.String("a1", "1"),
		slog.Int("a2", 2),
		slog.Group("sub", slog.Bool("b", true)),
	))

	l.With("number", 56).Log(slog.InfoLevel+1, "hello", "who", "world!")

	l.WithGroup("XXXXX").
		With("http", "HTTP").
		WithGroup("YYYYY").
		Info("yolo", "swag", 42)

	// Ignored because the level is under Info
	l.Debug("yoloDebug", "swag", 42)

	assert.Equal(t, []string{
		"WARN [simple] attr=42",
		"ERROR [fatal] group.a1=1 group.a2=2 group.sub.b=true err=EOF",
		"INFO+1 [hello] number=56 who=world!",
		"INFO [yolo] XXXXX.http=HTTP XXXXX.YYYYY.swag=42",
	}, h.Records())
}
