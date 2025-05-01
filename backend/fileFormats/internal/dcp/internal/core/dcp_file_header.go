package dcpCore

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/dcpParts"
	"fmt"
	"io"
	"os"
)

type (
	dataRange struct {
		Start int64
		End   int64
	}

	dataOffset struct {
		Offset int64
		Value  uint32
	}

	IDcpFileHeader interface {
		FromFile(file string) error
		DataLengths(header *dcpFileHeader, file *os.File) error
		Update(dcpParts components.IList[dcpParts.DcpFileParts]) error
		Write(buffer *bytes.Buffer) error
	}

	dcpFileHeader struct {
		Header     [0x40]byte
		Pointers   []dataOffset
		DataRanges []dataRange
	}
)

func newHeader() *dcpFileHeader {
	return &dcpFileHeader{
		Pointers: make([]dataOffset, 0, 7),
	}
}

func (h *dcpFileHeader) GetHeader() [0x40]byte {
	return h.Header
}

func (h *dcpFileHeader) FromFile(file string) error {
	openFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("error when opening the file: %s", file)
	}

	if _, err := io.ReadFull(openFile, h.Header[:]); err != nil {
		return fmt.Errorf("error reading the header from file: %s", file)
	}

	if err := h.getPointers(); err != nil {
		return err
	}

	return nil
}

func (h *dcpFileHeader) DataLengths(header *dcpFileHeader, file *os.File) error {
	createDataRanges := func(index int, count int, data []dataOffset) error {
		ranges := dataRange{}
		ranges.Start = int64(data[index].Value)

		if next := index + 1; next < count {
			ranges.End = int64(header.Pointers[next].Value)
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				return fmt.Errorf("error getting file info: %s", file.Name())
			}

			ranges.End = fileInfo.Size()
		}

		h.DataRanges = append(h.DataRanges, ranges)

		return nil
	}

	for i := 0; i < len(header.Pointers); i++ {
		if err := createDataRanges(i, len(header.Pointers), header.Pointers); err != nil {
			return err
		}
	}

	return nil
}

func (h *dcpFileHeader) Update(dcpParts components.IList[dcpParts.DcpFileParts]) error {
	var currentOffset = uint32(h.Pointers[0].Value)
	items := dcpParts.GetItems()

	for i, pointer := range h.Pointers {
		partInfo, err := os.Stat(items[i].GetDestination().Import().GetTargetFile())
		if err != nil {
			return fmt.Errorf("error getting file info: %s", items[i].GetDestination().Import().GetTargetFile())
		}

		if i == 0 {
			currentOffset = uint32(pointer.Value) + uint32(partInfo.Size())
			continue
		}

		newPointer := currentOffset
		binary.LittleEndian.PutUint32(h.Header[pointer.Offset:], newPointer)

		currentOffset = newPointer + uint32(partInfo.Size())
	}

	return nil
}

func (h *dcpFileHeader) Write(buffer *bytes.Buffer) error {
	if _, err := buffer.Write(h.Header[:]); err != nil {
		return fmt.Errorf("error when writing the header: %s", err.Error())
	}

	return nil
}

func (h *dcpFileHeader) getPointers() error {
	for i := 0; i < len(h.Header); i += 4 {
		value := binary.LittleEndian.Uint32(h.Header[i : i+4])

		if value != 0 {
			h.Pointers = append(h.Pointers, dataOffset{
				Offset: int64(i),
				Value:  value,
			})
		}
	}

	return nil
}
