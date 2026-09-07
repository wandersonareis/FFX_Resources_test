package models

import (
	"encoding/binary"
	"fmt"
	"io"
)

type (
	Offset uint16
	Key    uint16

	Segment struct {
		Offset Offset
		Key    Key
	}

	NameDescriptionHeaderData struct {
		NameSegment               Segment
		SimplifiedNameSegment     Segment
		DescriptionSegment        Segment
		SimplifiedDescriptionSegment Segment
	}

	NameOnlyHeaderData struct {
		NameSegment           Segment
		FirstSeparatorSegment Segment
	}
)

func ReadSegment(reader io.Reader) (Segment, error) {
	var offset, key uint16
	if err := binary.Read(reader, binary.LittleEndian, &offset); err != nil {
		if err == io.EOF {
			panic("Unexpected EOF while reading segment offset")
		}
		return Segment{}, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &key); err != nil {
		if err == io.EOF {
			panic("Unexpected EOF while reading segment key")
		}
		return Segment{}, err
	}
	return Segment{Offset: Offset(offset), Key: Key(key)}, nil
}

func WriteSegment(writer io.Writer, s Segment) error {
	err := binary.Write(writer, binary.LittleEndian, uint16(s.Offset))
	if err != nil {
		return err
	}
	return binary.Write(writer, binary.LittleEndian, uint16(s.Key))
}

func WriteSegmentAt(data []byte, pos int, s Segment) error {
    if pos < 0 || pos+4 > len(data) {
        return fmt.Errorf("invalid position %d for segment (slice length: %d)", pos, len(data))
    }
    binary.LittleEndian.PutUint16(data[pos:pos+2], uint16(s.Offset))
    binary.LittleEndian.PutUint16(data[pos+2:pos+4], uint16(s.Key))
    return nil
}
