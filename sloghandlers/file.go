package sloghandlers

import (
	"compress/gzip"
	"golang.org/x/exp/slog"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Create a new slog.Handler that store record in a file (base/log/2006-01-02+150405.log.gz).
// Format is a JSON objet for each record on a line, an compressed by gzip.
func NewFileHandler(base string, level slog.Level) (func() error, slog.Handler) {
	return newFileHandler(base, level, time.Now())
}

func newFileHandler(base string, level slog.Level, t time.Time) (func() error, slog.Handler) {
	if err := os.MkdirAll(filepath.Join(base, "log"), 0o775); err != nil {
		log.Fatal("Create log directory fail:", err)
	}

	file, err := os.Create(filepath.Join(base, "log", t.UTC().Format("2006-01-02+150405.log.gz")))
	if err != nil {
		log.Fatal("Create log file fail:", err)
	}

	compressor := gzip.NewWriter(file)

	closer := func() error {
		if err := compressor.Close(); err != nil {
			return err
		}
		return file.Close()
	}

	return closer, slog.HandlerOptions{Level: level}.NewJSONHandler(compressor)
}
