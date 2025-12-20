package models

import (
	"encoding/binary"
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
		NameSegment            Segment
		FirstSeparatorSegment  Segment
		DescriptionSegment     Segment
		SecondSeparatorSegment Segment
	}

	NameOnlyHeaderData struct {
		NameSegment           Segment
		FirstSeparatorSegment Segment
	}
)

func ReadSegment(reader io.Reader) (Segment, error) {
	var offset, key uint16
	err := binary.Read(reader, binary.LittleEndian, &offset)
	if err != nil {
		return Segment{}, err
	}
	err = binary.Read(reader, binary.LittleEndian, &key)
	if err != nil {
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
