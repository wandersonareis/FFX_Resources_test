package file

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/core/components"
	"ffxresources/backend/fileFormats/internal/dcp/internal/parts"
	"ffxresources/backend/logger"
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
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

	log zerolog.Logger
}

func NewHeader() *Header {
	return &Header{
		Pointers: make([]Pointer, 0, 7),

		log: logger.Get().With().Str("module", "dcp_file_header").Logger(),
	}
}

func (h *Header) GetHeader() [0x40]byte {
	return h.Header
}

func (h *Header) FromFile(file string) error {
	openFile, err := os.Open(file)
	if err != nil {
		h.log.Error().
			Err(err).
			Str("file", file).
			Msg("error when opening the file")

		return fmt.Errorf("error when opening the file")
	}

	if _, err := io.ReadFull(openFile, h.Header[:]); err != nil {
		h.log.Error().
			Err(err).
			Str("file", file).
			Msg("error reading the header")

		return fmt.Errorf("error reading the header")
	}

	if err := h.getPointers(); err != nil {
		return err
	}

	return nil
}

func (h *Header) DataLengths(header *Header, file *os.File) error {
	createDataRanges := func(index int, count int, data []Pointer) error {
		ranges := DataLength{}
		ranges.Start = int64(data[index].Value)

		if next := index + 1; next < count {
			ranges.End = int64(header.Pointers[next].Value)
		} else {
			fileInfo, err := file.Stat()
			if err != nil {
				h.log.Error().
					Err(err).
					Str("file", file.Name()).
					Msg("error getting file info")

				return err
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

func (h *Header) Update(dcpParts components.IList[parts.DcpFileParts]) error {
	var currentOffset = uint32(h.Pointers[0].Value)
	items := dcpParts.GetItems()

	for i, pointer := range h.Pointers {
		partInfo, err := os.Stat(items[i].GetImportLocation().TargetFile)
		if err != nil {
			h.log.Error().
				Err(err).
				Str("file", items[i].GetImportLocation().TargetFile).
				Msg("error getting file info")

			return err
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

func (h *Header) Write(buffer *bytes.Buffer) error {
	if _, err := buffer.Write(h.Header[:]); err != nil {
		h.log.Error().
			Err(err).
			Msg("error when writing the header")

		return err
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
