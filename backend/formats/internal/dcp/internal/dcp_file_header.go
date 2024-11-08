package internal

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"fmt"
	"io"
	"os"
	"reflect"
)

type DataLength struct {
	Start int64
	End   int64
}

type Pointer struct {
	Offset int64
	Value  uint32
}

type Header struct {
	Header     [0x40]byte
	Pointers   []Pointer
	DataRanges []DataLength
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

func (h *Header) DataLengths(header *Header, file *os.File) error {
	worker := common.NewWorker[Pointer]()

	worker.ForIndex(header.Pointers,
	func(index int, count int, data []Pointer) error {
		ranges := DataLength{}
		ranges.Start = int64(data[index].Value)

		if next := index+1; next < count {
			ranges.End = int64(header.Pointers[next].Value)
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				return err
			}
			ranges.End = fileInfo.Size()
		}

		h.DataRanges = append(h.DataRanges, ranges)

		return nil
	})

	results := h.DataRanges
	h.DataRanges = []DataLength{}
	
	for i := 0; i < len(header.Pointers); i++ {
		ranges := DataLength{}
		ranges.Start = int64(header.Pointers[i].Value)

		if i+1 < len(header.Pointers) {
			ranges.End = int64(header.Pointers[i+1].Value)
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				return fmt.Errorf("erro ao obter informações do arquivo: %w", err)
			}
			ranges.End = fileInfo.Size()
		}

		h.DataRanges = append(h.DataRanges, ranges)
	}

	if reflect.DeepEqual(h.DataRanges, results) {
		fmt.Println("ranges calculated successfully")
	} else {
		fmt.Println("error calculating the ranges")
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
