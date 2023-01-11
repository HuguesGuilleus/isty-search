package indexdatabase

type Bytes []byte

func (bytes Bytes) MarshalBinary() ([]byte, error) { return []byte(bytes), nil }
func UnmarshalBytes(data []byte) (Bytes, error) {
	b := make([]byte, len(data))
	copy(b, data)
	return b, nil
}
