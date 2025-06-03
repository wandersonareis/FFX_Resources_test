package components

type Chunk struct {
	Bytes  []byte
	Offset int
	Length int
}

// Chunk vazio
func NewEmptyChunk() Chunk {
	return Chunk{
		Bytes:  []byte{},
		Offset: 0,
		Length: 0,
	}
}

// Chunk com dados copiados entre os offsets, com proteção de bounds
func NewChunk(data []byte, from, to int) Chunk {
	if from >= len(data) {
		return Chunk{
			Bytes:  []byte{},
			Offset: from,
			Length: to - from,
		}
	}
	if to > len(data) {
		to = len(data)
	}
	bytes := make([]byte, to-from)
	copy(bytes, data[from:to])
	return Chunk{
		Bytes:  bytes,
		Offset: from,
		Length: to - from,
	}
}
