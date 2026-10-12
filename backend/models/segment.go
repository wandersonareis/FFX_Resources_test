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
		NameSegment                  Segment
		SimplifiedNameSegment        Segment
		DescriptionSegment           Segment
		SimplifiedDescriptionSegment Segment
	}

	NameOnlyHeaderData struct {
		NameSegment           Segment
		FirstSeparatorSegment Segment
	}
)

// TextPart identifies one reference contained in a text field. TextRef uses
// Standard; TextPair uses both Standard and Simplified.
type TextPart string

const (
	TextPartStandard   TextPart = "Standard"
	TextPartSimplified TextPart = "Simplified"
)

// TextValue is one text reference embedded in a binary schema field.
type TextValue struct {
	Part    TextPart
	Segment Segment
	// SameAs declares the localization fallback relationship for this part.
	// If omitted during export, its source part supplies the translated text.
	SameAs TextPart
}

// TextSegmentsExporter and TextSegmentsImporter deliberately describe
// separate capabilities: extraction does not require write access, and
// importing text does not require exposing the schema's binary representation.
type (
	TextSegmentsExporter interface {
		ExportTextSegments() []TextValue
	}
	TextSegmentsImporter interface {
		ImportTextSegments([]TextValue) error
	}
)

// TextRef is the chunk-mapping flow's explicit text-reference type. Segment
// remains unchanged for the existing objectsfile flow.
type TextRef Segment

var (
	_ TextSegmentsExporter = TextRef{}
	_ TextSegmentsImporter = (*TextRef)(nil)
)

func (r TextRef) ExportTextSegments() []TextValue {
	return []TextValue{{Part: TextPartStandard, Segment: Segment(r)}}
}

func (r *TextRef) ImportTextSegments(values []TextValue) error {
	if len(values) != 1 || values[0].Part != TextPartStandard {
		return fmt.Errorf("TextRef expects one Standard text segment")
	}
	*r = TextRef(values[0].Segment)
	return nil
}

// SimplifiedRef é a referência simplificada isolada: exporta Part Simplified
// sem SameAs — para quando o par Standard/Simplified NÃO é adjacente no
// chunk (arrays separados, como names/names_hira de w_name.bin), onde o
// fallback por SameAs do TextPair não alcança. O relacionamento entre as
// partes fica a cargo do exportador/importador do domínio.
type SimplifiedRef Segment

var (
	_ TextSegmentsExporter = SimplifiedRef{}
	_ TextSegmentsImporter = (*SimplifiedRef)(nil)
)

func (r SimplifiedRef) ExportTextSegments() []TextValue {
	return []TextValue{{Part: TextPartSimplified, Segment: Segment(r)}}
}

func (r *SimplifiedRef) ImportTextSegments(values []TextValue) error {
	if len(values) != 1 || values[0].Part != TextPartSimplified {
		return fmt.Errorf("SimplifiedRef expects one Simplified text segment")
	}
	*r = SimplifiedRef(values[0].Segment)
	return nil
}

// TextPair knows that both references are text, preserves their semantic
// order, and declares that Simplified shares Standard unless it differs.
type TextPair struct {
	Standard   TextRef
	Simplified TextRef
}

var (
	_ TextSegmentsExporter = TextPair{}
	_ TextSegmentsImporter = (*TextPair)(nil)
)

func (p TextPair) ExportTextSegments() []TextValue {
	return []TextValue{
		{Part: TextPartStandard, Segment: Segment(p.Standard)},
		{Part: TextPartSimplified, Segment: Segment(p.Simplified), SameAs: TextPartStandard},
	}
}

func (p *TextPair) ImportTextSegments(values []TextValue) error {
	if len(values) != 2 || values[0].Part != TextPartStandard || values[1].Part != TextPartSimplified {
		return fmt.Errorf("TextPair expects Standard and Simplified text segments")
	}
	p.Standard = TextRef(values[0].Segment)
	p.Simplified = TextRef(values[1].Segment)
	return nil
}

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

func ReadSegmentAt(data []byte, pos int) (Segment, error) {
	if pos < 0 || pos+4 > len(data) {
		return Segment{}, fmt.Errorf("invalid position %d for segment (slice length: %d)", pos, len(data))
	}
	return Segment{
		Offset: Offset(binary.LittleEndian.Uint16(data[pos : pos+2])),
		Key:    Key(binary.LittleEndian.Uint16(data[pos+2 : pos+4])),
	}, nil
}

func WriteSegmentAt(data []byte, pos int, s Segment) error {
	if pos < 0 || pos+4 > len(data) {
		return fmt.Errorf("invalid position %d for segment (slice length: %d)", pos, len(data))
	}
	binary.LittleEndian.PutUint16(data[pos:pos+2], uint16(s.Offset))
	binary.LittleEndian.PutUint16(data[pos+2:pos+4], uint16(s.Key))
	return nil
}
