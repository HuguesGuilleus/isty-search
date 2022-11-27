package sloghandlers

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"testing"
)

func TestHandlerSlice(t *testing.T) {
	h1 := NewHandlerRecords(slog.InfoLevel)
	h2 := NewHandlerRecords(slog.InfoLevel)
	fillLogger(NewMultiHandlers(h1, h2))

	assert.Equal(t, expectedRecords, h1.Records())
	assert.Equal(t, expectedRecords, h2.Records())
}
