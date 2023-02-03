package indexdatabase

import (
	"bytes"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/keys"
	"os"
)

func Store[T any](file string, m map[keys.Key]T, marshaler func(T) []byte) error {
	buff := bytes.Buffer{}

	for key, value := range m {
		data := marshaler(value)
		buff.Write(key[:])
		buff.WriteByte(byte(len(data) >> 24))
		buff.WriteByte(byte(len(data) >> 16))
		buff.WriteByte(byte(len(data) >> 8))
		buff.WriteByte(byte(len(data) >> 0))
		buff.Write(data)
	}

	return os.WriteFile(file, buff.Bytes(), 0o664)
}

func Load[T any](file string, unmarshal func([]byte) (T, error)) (map[keys.Key]T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Load(%q): %w", file, err)
	}

	m := make(map[keys.Key]T)
	for i := 0; i+keys.Len+4 < len(data); {
		// Get the key
		key := keys.Key{}
		copy(key[:], data[i:])
		i += keys.Len

		// Get the length
		l := int(data[i+0])<<24 |
			int(data[i+1])<<16 |
			int(data[i+2])<<8 |
			int(data[i+3])<<0
		i += 4

		if i+l > len(data) {
			return nil, fmt.Errorf("The length of key %s is too long fo the current file length", key)
		}

		// Unmarshal the data
		value, err := unmarshal(data[i : i+l])
		if err != nil {
			return nil, fmt.Errorf("Unmarshal value of %s: %w", key, err)
		}
		i += l

		// Store the value
		m[key] = value
	}

	return m, nil
}
