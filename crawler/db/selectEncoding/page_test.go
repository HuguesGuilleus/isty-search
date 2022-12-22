package choseEncoding

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"github.com/HuguesGuilleus/isty-search/common"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"io"
	"testing"
	"time"
)

var (
	//go:embed Nicéphore_II_Phocas.html
	testPageSourceBytes []byte
	testPageSource      = func() *crawler.Page {
		pageSourceURL := common.ParseURL("https://fr.wikipedia.org/wiki/Wikip%C3%A9dia:Accueil_principal")
		node, err := htmlnode.Parse(testPageSourceBytes)
		if err != nil {
			panic(err)
		}

		return &crawler.Page{
			URL:  *pageSourceURL,
			Time: time.Now().UTC(),
			Html: node,
		}
	}()

	encoderSlice = []struct {
		name   string
		encode func(any) ([]byte, error)
		decode func([]byte, any) error
	}{
		{"gob", func(v any) ([]byte, error) {
			buff := bytes.Buffer{}
			err := gob.NewEncoder(&buff).Encode(v)
			return buff.Bytes(), err
		}, func(data []byte, v any) error {
			return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
		}},
		{"json", json.Marshal, json.Unmarshal},
	}

	compressorSlice = []struct {
		name         string
		compressor   func(io.Writer) io.WriteCloser
		decompressor func(io.Reader) (io.ReadCloser, error)
	}{
		{"none", nil, nil},
		{"gzip",
			func(w io.Writer) io.WriteCloser { return gzip.NewWriter(w) },
			func(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) }},
		{"zlib",
			func(w io.Writer) io.WriteCloser { return zlib.NewWriter(w) },
			zlib.NewReader},
		{"zS__",
			func(w io.Writer) io.WriteCloser { return newWriterPanic(zlib.NewWriterLevel(w, zlib.BestSpeed)) },
			zlib.NewReader},
		{"zC__",
			func(w io.Writer) io.WriteCloser { return newWriterPanic(zlib.NewWriterLevel(w, zlib.BestCompression)) },
			zlib.NewReader},
	}
)

func newWriterPanic(w io.WriteCloser, err error) io.WriteCloser {
	if err != nil {
		panic(err)
	}
	return w
}

func TestSaver(t *testing.T) {
	format := "%-10s: %10d B"

	// Base HTML
	for _, compressor := range compressorSlice {
		data := testEncoder(nil, fakeEncoder, compressor.compressor)
		t.Logf(format, "html+"+compressor.name, len(data))
	}
	t.Log()

	// crawler.Page
	for _, encoder := range encoderSlice {
		for _, compressor := range compressorSlice {
			name := encoder.name + "+" + compressor.name
			data := testEncoder(testPageSource, encoder.encode, compressor.compressor)
			page := testDecoder[crawler.Page](data, encoder.decode, compressor.decompressor)
			assert.Equal(t, testPageSource.URL, page.URL, name)
			assert.Equal(t, testPageSource.Time, page.Time, name)
			assert.Equal(t, testPageSource.Html.RootId, page.Html.RootId, name)
			assert.Equal(t, testPageSource.Html.RootClasses, page.Html.RootClasses, name)
			assert.Equal(t, testPageSource.Html.RootAttributes, page.Html.RootAttributes, name)
			assert.Equal(t, testPageSource.Html.Meta, page.Html.Meta, name)
			assert.Equal(t, testPageSource.Html.Head.PrintLines(), page.Html.Head.PrintLines(), name)
			assert.Equal(t, testPageSource.Html.Body.PrintLines(), page.Html.Body.PrintLines(), name)
			t.Logf(format, name, len(data))
		}
		t.Log()
	}
}

func BenchmarkEncode(b *testing.B) {
	for _, compressor := range compressorSlice {
		b.Run("html+"+compressor.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				testEncoder(nil, fakeEncoder, compressor.compressor)
			}
		})
	}

	for _, encoder := range encoderSlice {
		for _, compressor := range compressorSlice {
			name := encoder.name + "+" + compressor.name
			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					testEncoder(testPageSource, encoder.encode, compressor.compressor)
				}
			})
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	for _, compressor := range compressorSlice {
		data := testEncoder(nil, fakeEncoder, compressor.compressor)

		b.Run("html+"+compressor.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buff := &bytes.Buffer{}
				if compressor.decompressor != nil {
					decompress, _ := compressor.decompressor(bytes.NewReader(data))
					buff.ReadFrom(decompress)
					decompress.Close()
				} else {
					buff = bytes.NewBuffer(data)
				}
				html.Parse(buff)
			}
		})
	}

	for _, encoder := range encoderSlice {
		for _, compressor := range compressorSlice {
			name := encoder.name + "+" + compressor.name
			data := testEncoder(testPageSource, encoder.encode, compressor.compressor)

			b.Run(name, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					testDecoder[crawler.Page](data, encoder.decode, compressor.decompressor)
				}
			})
		}
	}
}

// Return always testPageSourceBytes, nil
func fakeEncoder(_ any) ([]byte, error) { return testPageSourceBytes, nil }

func testEncoder(v any, encode func(any) ([]byte, error), compressor func(io.Writer) io.WriteCloser) []byte {
	data, err := encode(v)
	if err != nil {
		panic(err)
	}

	if compressor == nil {
		return data
	}

	buff := bytes.Buffer{}
	c := compressor(&buff)

	if n, err := io.Copy(c, bytes.NewReader(data)); err != nil {
		panic(err)
	} else if int(n) != len(data) {
		panic("no write all")
	}
	if err := c.Close(); err != nil {
		panic(err)
	}
	return buff.Bytes()
}
func testDecoder[T any](data []byte, decode func([]byte, any) error, decompressor func(io.Reader) (io.ReadCloser, error)) *T {
	if decompressor != nil {
		decompress, err := decompressor(bytes.NewReader(data))
		if err != nil {
			panic(err)
		}
		buff := bytes.Buffer{}
		if _, err = buff.ReadFrom(decompress); err != nil {
			panic(err)
		}
		if err := decompress.Close(); err != nil {
			panic(err)
		}
		data = buff.Bytes()
	}

	v := new(T)
	if err := decode(data, v); err != nil {
		panic(err)
	}
	return v
}
