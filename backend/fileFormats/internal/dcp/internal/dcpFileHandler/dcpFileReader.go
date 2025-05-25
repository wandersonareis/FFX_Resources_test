package dcpFileHandler

import (
	"encoding/binary"
	"fmt"

	"ffxresources/backend/common"
)

type (
	Chunk struct {
		Data   []byte
		Offset int
		Length int
	}
	dcpFileReader struct{}
)

func newDcpFileReader() *dcpFileReader {
	return &dcpFileReader{}
}

func readUint32(b []byte, offset int) (uint32, error) {
	if b == nil || offset+4 > len(b) {
		return 0, fmt.Errorf("insufficient bytes to read uint32")
	}
	return binary.LittleEndian.Uint32(b[offset:]), nil
}

func (h *dcpFileReader) GetChunks(data []byte) ([]*Chunk, error) {
	if err := common.CheckArgumentNil(data, "data"); err != nil {
		return nil, err
	}

	chunkStartOffset := 0
	chunkCount := 16
	offsets := make([]uint32, chunkCount)

	for i := range offsets {
		offset, err := readUint32(data, i*4+chunkStartOffset)
		if err != nil {
			return nil, err
		}
		if offset == 0xFFFFFFFF {
			chunkCount = i - 1
			break
		}
		offsets[i] = offset
	}

	chunks := make([]*Chunk, 0, chunkCount)
	for i := range offsets {
		curOffset := offsets[i]
		if curOffset == 0 {
			chunks = append(chunks, &Chunk{})
			continue
		}
		to := uint32(len(data))
		for j := i + 1; j <= chunkCount; j++ {
			if j >= len(offsets) {
				break
			}
			if offsets[j] >= curOffset {
				to = offsets[j]
				break
			}
		}
		chunkData := make([]byte, to-curOffset)
		copy(chunkData, data[curOffset:to])
		chunks = append(chunks, &Chunk{
			Data:   chunkData,
			Offset: int(curOffset),
			Length: int(to - curOffset),
		})
	}

	return chunks, nil
}
