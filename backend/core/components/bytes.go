package components

import (
	"encoding/binary"
	"fmt"
	"os"
)

func read4BytesLE(data []byte, offset int) uint32 {
	if offset+4 > len(data) {
		return 0xFFFFFFFF // fallback se o slice for inválido
	}
	return binary.LittleEndian.Uint32(data[offset : offset+4])
}

func write4Bytes(bytes []byte, offset int, value uint32) {
	binary.LittleEndian.PutUint32(bytes[offset:], value)
}

func FileToBytes(path string, print bool) []byte {
	filePath, err := ResolveFile(path, print)
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return data
}

func BytesToChunks(data []byte, assumedChunkCount, chunkOffset int) []Chunk {
	if data == nil {
		return nil
	}

	chunkCount := assumedChunkCount
	offsets := make([]int, chunkCount+1)

	for i := 0; i < chunkCount; i++ {
		offset := int(read4BytesLE(data, i*4+chunkOffset))
		if offset == 0xFFFFFFFF {
			chunkCount = i - 1
			break
		}
		offsets[i] = offset
	}

	chunks := make([]Chunk, 0, chunkCount)
	for i := 0; i < chunkCount; i++ {
		offset := offsets[i]
		if offset == 0 {
			chunks = append(chunks, NewEmptyChunk())
		} else {
			to := len(data)
			for j := i + 1; j <= chunkCount; j++ {
				if offsets[j] >= offset {
					to = offsets[j]
					break
				}
			}

			chunks = append(chunks, NewChunk(data, offset, to))
		}
	}
	return chunks
}

func ChunksToBytes(chunks [][]byte, chunkCount, chunkInitialOffset, chunkAlignment int) ([]byte, error) {
	if len(chunks) > (chunkInitialOffset-0x08)/0x04 {
		return nil, fmt.Errorf("too many chunks for initial offset")
	}

	header := make([]byte, chunkInitialOffset)
	terminateWithFFs := false

	if chunkCount < 0 {
		write4Bytes(header, 0x00, 0x31305645) // escreve "EV01" em hex
		chunkCount = len(chunks)
		terminateWithFFs = true
	} else {
		write4Bytes(header, 0x00, uint32(chunkCount))
		chunkCount--
	}

	endOffset := chunkInitialOffset
	paddings := make([]int, chunkCount)

	for i := 0; i < chunkCount; i++ {
		addressTargetOffset := 0x04 + 0x04*i
		chunk := chunks[i]
		padding := 0

		if len(chunk) == 0 {
			write4Bytes(header, addressTargetOffset, 0)
		} else {
			write4Bytes(header, addressTargetOffset, uint32(endOffset))
			endOffset += len(chunk)

			if chunkAlignment > 1 {
				misalignment := endOffset % chunkAlignment
				if misalignment > 0 {
					padding = chunkAlignment - misalignment
					endOffset += padding
				}
			}
		}

		paddings[i] = padding
	}

	chunkListEndOffset := 0x04 + 0x04*chunkCount
	write4Bytes(header, chunkListEndOffset, uint32(endOffset))

	if terminateWithFFs {
		write4Bytes(header, chunkListEndOffset+0x04, 0xFFFFFFFF)
	}

	// Construir o slice completo (header + chunks + paddings)
	fullBytes := make([]byte, endOffset)
	copy(fullBytes, header)

	offset := chunkInitialOffset
	for i := 0; i < chunkCount; i++ {
		chunk := chunks[i]
		if len(chunk) > 0 {
			copy(fullBytes[offset:], chunk)
			offset += len(chunk)
		}
		if paddings[i] > 0 {
			for j := 0; j < paddings[i]; j++ {
				fullBytes[offset] = 0
				offset++
			}
		}
	}

	return fullBytes, nil
}
