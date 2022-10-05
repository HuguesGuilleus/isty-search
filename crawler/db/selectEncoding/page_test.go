package choseEncoding

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	_ "embed"
	"encoding/gob"
	"encoding/json"
	"github.com/HuguesGuilleus/isty-search/crawler"
	"github.com/HuguesGuilleus/isty-search/crawler/htmlnode"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"testing"
	"time"
)

type Compressor interface {
	io.Writer
	Flush() error
}

var (
	//go:embed Nicéphore_II_Phocas.html
	testPageSourceBytes []byte
	testPageSource      = func() *crawler.Page {
		pageSourceURL, err := url.Parse("https://fr.wikipedia.org/wiki/Wikip%C3%A9dia:Accueil_principal")
		if err != nil {
			panic(err)
		}
		node, err := htmlnode.Parse(testPageSourceBytes)
		if err != nil {
			panic(err)
		}

		return &crawler.Page{
			URL:  *pageSourceURL,
			Time: time.Now().UTC(),
			Node: node,
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
		compressor   func(io.Writer) Compressor
		decompressor func(io.Reader) (io.ReadCloser, error)
	}{
		{"none", nil, nil},
		{"gzip",
			func(w io.Writer) Compressor { return gzip.NewWriter(w) },
			func(r io.Reader) (io.ReadCloser, error) { return gzip.NewReader(r) }},
		{"zlib",
			func(w io.Writer) Compressor { return zlib.NewWriter(w) },
			zlib.NewReader},
	}
)

func TestSaver(t *testing.T) {
	format := "%-10s: %10d B"

	// Base HTML
	for _, compressor := range compressorSlice {
		data := testEncoder(nil, func(_ any) ([]byte, error) { return testPageSourceBytes, nil }, compressor.compressor)
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
			assert.Equal(t, testPageSource.Node.PrintLines(), page.Node.PrintLines(), name)
			t.Logf(format, name, len(data))
		}
		t.Log()
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

func testEncoder(v any, encode func(any) ([]byte, error), compressor func(io.Writer) Compressor) []byte {
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
	if err := c.Flush(); err != nil {
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
		buff.ReadFrom(decompress)
		decompress.Close()
		data = buff.Bytes()
	}

	v := new(T)
	if err := decode(data, v); err != nil {
		panic(err)
	}
	return v
}
