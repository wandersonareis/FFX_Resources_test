package dcpFileHandler

import (
	"encoding/binary"
	"fmt"
	"os"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
)

type (
	dcpFileWriter struct{}
)

func newDcpFileWriter() *dcpFileWriter {
	return &dcpFileWriter{}
}

func (h *dcpFileWriter) UpdateChunks(originalChunks []*Chunk, parts components.IList[dcpParts.DcpFileParts]) ([]*Chunk, error) {
	partsItems := parts.GetItems()
	if len(partsItems) == 0 {
		return nil, fmt.Errorf("no parts provided")
	}

	updatedChunks := make([]*Chunk, len(originalChunks))
	var fileIndex int
	var nextOffset int
	firstChunkProcessed := false

	for i, chunk := range originalChunks {
		if chunk.Offset == 0 {
			updatedChunks[i] = chunk
			continue
		}

		if fileIndex >= len(partsItems) {
			return nil, fmt.Errorf("file index out of range: %d out of %d", fileIndex, len(partsItems))
		}

		fileData, err := h.readFileData(partsItems[fileIndex])
		if err != nil {
			return nil, err
		}

		if !firstChunkProcessed {
			nextOffset = chunk.Offset
			firstChunkProcessed = true
		}

		updatedChunks[i] = &Chunk{
			Data:   fileData,
			Offset: nextOffset,
			Length: len(fileData),
		}

		nextOffset += len(fileData)
		fileIndex++
	}

	return updatedChunks, nil
}

func (h *dcpFileWriter) BuildHeader(chunks []*Chunk) []byte {
	const pointerSize = 4
	offsets := make([]uint32, len(chunks)+1)

	currentOffset := uint32(pointerSize * (len(chunks) + 1))
	for i, chunk := range chunks {
		if len(chunk.Data) == 0 {
			offsets[i] = 0
		} else {
			offsets[i] = currentOffset
			currentOffset += uint32(len(chunk.Data))
		}
	}
	offsets[len(chunks)] = currentOffset

	headerBytes := make([]byte, pointerSize*len(offsets))
	for i, off := range offsets {
		binary.LittleEndian.PutUint32(headerBytes[i*4:], off)
	}
	return headerBytes
}

func (h *dcpFileWriter) WriteFile(targetPath string, data []byte) error {
	if err := common.WriteBytesToFile(targetPath, data); err != nil {
		return fmt.Errorf("error saving the file: %s", targetPath)
	}
	return nil
}
func (h *dcpFileWriter) readFileData(part dcpParts.DcpFileParts) ([]byte, error) {
	targetFile := part.GetDestination().Import().GetTargetFile()
	data, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", targetFile, err)
	}
	return data, nil
}

func (h *dcpFileWriter) SaveContainerFile(path string, chunks []*Chunk) error {
	header := h.BuildHeader(chunks)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo: %w", err)
	}
	defer file.Close()

	_, err = file.Write(header)
	if err != nil {
		return fmt.Errorf("erro ao escrever header: %w", err)
	}

	for _, chunk := range chunks {
		if chunk.Data == nil {
			continue
		}
		_, err := file.Write(chunk.Data)
		if err != nil {
			return fmt.Errorf("erro ao escrever chunk: %w", err)
		}
	}

	return nil
}
