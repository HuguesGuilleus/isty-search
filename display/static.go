package display

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"io/fs"
)

var (
	//go:embed css
	cssFS    embed.FS
	cssBytes = concatMinify(cssFS, css.Minify)
)

func concatMinify(fsys fs.FS, minifier minify.MinifierFunc) []byte {
	buff := bytes.Buffer{}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("Read %q: %w", path, err)
		}

		if err := minifier(nil, &buff, bytes.NewReader(data), nil); err != nil {
			return fmt.Errorf("Minify %q: %w", path, err)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return buff.Bytes()
}
