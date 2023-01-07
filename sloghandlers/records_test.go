package sloghandlers

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"io"
	"testing"
)

var expectedRecords = []string{
	"WARN [simple] attr=042 quoted=\"x=3\"",
	"ERROR [fatal] group.a1=A group.a2=+002 group.sub.b=true err=EOF",
	"INFO+1 [hello] number=-078_356 who=world!",
	"INFO [yolo] XXXXX.http=HTTP XXXXX.YYYYY.swag=+042",
}

func TestHandlerRecords(t *testing.T) {
	records, h := NewHandlerRecords(slog.InfoLevel)

	assert.False(t, h.Enabled(slog.DebugLevel))
	assert.True(t, h.Enabled(slog.InfoLevel))
	assert.True(t, h.Enabled(slog.WarnLevel))
	fillLogger(h)
	assert.Equal(t, expectedRecords, *records)
}

// Add record to the handler.
func fillLogger(h slog.Handler) {
	l := slog.New(h)
	l.Warn("simple", "attr", uint64(42), "quoted", `x=3`)
	l.Error("fatal", io.EOF, slog.Group("group",
		slog.String("a1", "A"),
		slog.Int("a2", 2),
		slog.Group("sub", slog.Bool("b", true)),
	))

	l.With("number", -78356).Log(slog.InfoLevel+1, "hello", "who", "world!")

	l.WithGroup("XXXXX").
		With("http", "HTTP").
		WithGroup("YYYYY").
		Info("yolo", "swag", 42)

	// Ignored because the level is under Info
	l.Debug("yoloDebug", "swag", 42)
}
