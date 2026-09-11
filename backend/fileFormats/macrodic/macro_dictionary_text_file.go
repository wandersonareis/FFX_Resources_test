package macrodic

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"fmt"
)

// MacroDictionaryTextSegment is a single macro string entry within a macro
// dictionary file: the Name/SimplifiedName offsets (uint16) into the file Bytes
// plus the resolved null-terminated text bytes. When both offsets are equal,
// both fields point to the same text (no distinct simplified text exists).
type MacroDictionaryTextSegment struct {
	NameOffset           uint16
	SimplifiedNameOffset uint16
	NameBytes            []byte
	SimplifiedNameBytes  []byte
}

// MacroDictionaryTextFile holds the raw bytes of a single file contained in a
// macro dictionary container, along with its parsed Name/SimplifiedName segments.
// It follows the MacroDictionaryBinaryFile/NameOnlyTextObject pattern: Bytes keeps the raw data,
// mapBytes parses it, ToBytes rebuilds it from zero. Nil segments are empty
// placeholders that rebuild as zero entries, preserving index alignment.
type MacroDictionaryTextFile struct {
	Bytes    []byte
	Segments []*MacroDictionaryTextSegment
	Charset  string
	Version  int
	// Index is the chunk index this file occupies in its container.
	// It is set by container FileAt/Files; files built outside a container
	// carry -1 until placed.
	Index int
}

// NewMacroDictionaryTextFile creates a text file from raw file bytes, parsing the
// segment offset table via mapBytes.
func NewMacroDictionaryTextFile(data []byte, charset string, version int) (*MacroDictionaryTextFile, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("insufficient data to create MacroDictionaryTextFile: have %d bytes, need at least 2", len(data))
	}

	f := &MacroDictionaryTextFile{
		Bytes:   data,
		Charset: charset,
		Version: version,
	}
	if err := f.mapBytes(); err != nil {
		return nil, err
	}
	return f, nil
}

// mapBytes parses the segment offset table at the start of the file Bytes.
// The segment count comes from MaxIndex (first offset / 4); each entry is a
// uint16 pair (Name offset + SimplifiedName offset). Text bytes are resolved
// with the same null-terminated lookup the legacy NewMacroString path uses,
// sharing the slice when both offsets are equal.
func (f *MacroDictionaryTextFile) mapBytes() error {
	count := MaxIndex(f.Bytes)
	if len(f.Bytes) < count*4 {
		return fmt.Errorf("insufficient data to map MacroDictionaryTextFile segments: have %d bytes, need at least %d", len(f.Bytes), count*4)
	}

	r := bytes.NewReader(f.Bytes)
	f.Segments = make([]*MacroDictionaryTextSegment, 0, count)
	for range count {
		var nameOff, simplifiedOff uint16
		if err := binary.Read(r, binary.LittleEndian, &nameOff); err != nil {
			return fmt.Errorf("reading name offset: %w", err)
		}
		if err := binary.Read(r, binary.LittleEndian, &simplifiedOff); err != nil {
			return fmt.Errorf("reading simplified name offset: %w", err)
		}

		nameBytes := common.GetStringBytesAtLookupOffset(f.Bytes, int(nameOff))
		var simplifiedBytes []byte
		if simplifiedOff == nameOff {
			simplifiedBytes = nameBytes
		} else {
			simplifiedBytes = common.GetStringBytesAtLookupOffset(f.Bytes, int(simplifiedOff))
		}

		seg := MacroDictionaryTextSegment{
			NameOffset:           nameOff,
			SimplifiedNameOffset: simplifiedOff,
			NameBytes:            nameBytes,
			SimplifiedNameBytes:  simplifiedBytes,
		}
		f.Segments = append(f.Segments, &seg)
	}
	return nil
}

// MaxIndex returns how many 4-byte segment entries this file holds,
// derived from its own Bytes via the first-offset / 4 formula.
func (f *MacroDictionaryTextFile) MaxIndex() int {
	return MaxIndex(f.Bytes)
}

// ToMacroStrings converts the file segments to MacroString objects,
// preserving positions (nil segments stay nil, like the legacy parse result).
// The returned structs share the file text byte slices (read-only use).
func (f *MacroDictionaryTextFile) ToMacroStrings() []*MacroString {
	if len(f.Segments) == 0 {
		return nil
	}
	result := make([]*MacroString, len(f.Segments))
	for i, seg := range f.Segments {
		if seg == nil {
			continue
		}
		result[i] = &MacroString{
			Charset:          f.Charset,
			Version:          f.Version,
			RegularOffset:    int(seg.NameOffset),
			SimplifiedOffset: int(seg.SimplifiedNameOffset),
			RegularBytes:     seg.NameBytes,
			SimplifiedBytes:  seg.SimplifiedNameBytes,
		}
	}
	return result
}

// ToBytes rebuilds the file binary from zero with RebuildChunkData (first
// offset = segment count * 4, offsets recalculated sequentially from the raw
// text bytes). It never alters the file Segments — changes only materialize in
// the returned slice, applied by the caller during saveToBinary execution.
func (f *MacroDictionaryTextFile) ToBytes() ([]byte, error) {
	if len(f.Segments) == 0 {
		return []byte{}, nil
	}
	return RebuildChunkData(f.Segments), nil
}
