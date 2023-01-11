package indexdatabase

import (
	"bytes"
	"encoding"
	"fmt"
	"github.com/HuguesGuilleus/isty-search/crawler/database"
	"os"
)

func Store[T encoding.BinaryMarshaler](file string, m map[crawldatabase.Key]T) error {
	buff := bytes.Buffer{}

	for key, value := range m {
		data, err := value.MarshalBinary()
		if err != nil {
			return fmt.Errorf("MarshalBinary for key %s: %w", key, err)
		}
		buff.Write(key[:])
		buff.WriteByte(byte(len(data) >> 24))
		buff.WriteByte(byte(len(data) >> 16))
		buff.WriteByte(byte(len(data) >> 8))
		buff.WriteByte(byte(len(data) >> 0))
		buff.Write(data)
	}

	return os.WriteFile(file, buff.Bytes(), 0o664)
}

func Load[T any](file string, unmarshal func([]byte) (T, error)) (map[crawldatabase.Key]T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Load(%q): %w", file, err)
	}

	m := make(map[crawldatabase.Key]T)
	for i := 0; i+crawldatabase.KeyLen+4 < len(data); {
		// Get the key
		key := crawldatabase.Key{}
		copy(key[:], data[i:])
		i += crawldatabase.KeyLen

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
