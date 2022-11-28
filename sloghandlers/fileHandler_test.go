package sloghandlers

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHandlerFile(t *testing.T) {
	defer os.RemoveAll("_test")

	now := time.Date(2009, 11, 13, 23, 0, 0, 0, time.UTC)
	r := slog.NewRecord(now, slog.InfoLevel, "yolo", 0, context.Background())
	r.AddAttrs(slog.Int("n", 42), slog.Bool("b", true), slog.String("hello", "world!"))

	closer, h := newFileHandler("_test", slog.InfoLevel, now)
	assert.NoError(t, h.Handle(r))
	assert.NoError(t, h.Handle(r))
	assert.NoError(t, closer())

	data, err := os.ReadFile(filepath.Join("_test", "log", "2009-11-13+230000.log.gz"))
	assert.NoError(t, err)

	reader, err := gzip.NewReader(bytes.NewReader(data))
	assert.NoError(t, err)
	decoder := json.NewDecoder(reader)

	for i := 0; i < 2; i++ {
		type recordType struct {
			Time    time.Time `json:"time"`
			Level   string    `json:"level"`
			Message string    `json:"msg"`
			N       int       `json:"n"`
			B       bool      `json:"b"`
			Hello   string    `json:"hello"`
		}
		jsonRecord := recordType{}
		assert.NoError(t, decoder.Decode(&jsonRecord))

		assert.Equal(t, recordType{
			Time:    now,
			Level:   "INFO",
			Message: "yolo",
			N:       42,
			B:       true,
			Hello:   "world!",
		}, jsonRecord, "i=", i)
	}
}
