package dcp_internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Pointer struct {
	Offset int64
	Value  uint32
}

type Header struct {
	Header   [0x40]byte
	Pointers []Pointer
}

func NewHeader() *Header {
	return &Header{
		Pointers: make([]Pointer, 0, 7),
	}
}

func (h *Header) GetHeader() [0x40]byte {
	return h.Header
}

func (h *Header) FromFile(file string) error {
	openFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("error when opening the file: %w", err)
	}

	if _, err := io.ReadFull(openFile, h.Header[:]); err != nil {
		return fmt.Errorf("error reading the header: %w", err)
	}

	if err := h.getPointers(); err != nil {
		return fmt.Errorf("error when getting the pointers: %w", err)
	}

	return nil
}

func (h *Header) Update(dcpParts []DcpFileParts) error {
	var currentOffset = uint32(h.Pointers[0].Value)

	for i, pointer := range h.Pointers {
		data := dcpParts[i].gameDataInfo.GameData

		if i == 0 {
			currentOffset = uint32(pointer.Value) + uint32(data.Size)
			continue
		}

		newPointer := currentOffset
		binary.LittleEndian.PutUint32(h.Header[pointer.Offset:], newPointer)

		currentOffset = newPointer + uint32(data.Size)
	}

	return nil
}

func (h *Header) Write(buffer *bytes.Buffer) error {
	if _, err := buffer.Write(h.Header[:]); err != nil {
		return fmt.Errorf("error when recording the header: %w", err)
	}

	return nil
}

func (h *Header) getPointers() error {
	for i := 0; i < len(h.Header); i += 4 {
		value := binary.LittleEndian.Uint32(h.Header[i : i+4])

		if value != 0 {
			h.Pointers = append(h.Pointers, Pointer{
				Offset: int64(i),
				Value:  value,
			})
		}
	}

	return nil
}
