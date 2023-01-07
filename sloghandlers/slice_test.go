package sloghandlers

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestHandlerSlice(t *testing.T) {
	r1, h1 := NewHandlerRecords(slog.DebugLevel + 1)
	r2, h2 := NewHandlerRecords(slog.InfoLevel)
	fillLogger(NewMultiHandlers(h1, h2))

	assert.Equal(t, expectedRecords, *r1)
	assert.Equal(t, expectedRecords, *r2)
}
