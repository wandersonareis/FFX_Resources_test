package event

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/models"
)

type FieldString struct {
	Charset           string
	Version           int
	RegularOffset     int
	RegularFlags      int
	RegularChoices    int
	SimplifiedOffset  int
	SimplifiedFlags   int
	SimplifiedChoices int
	RegularBytes      []byte
	SimplifiedBytes   []byte
}

func NewEmptyFieldString(charset string, version int) *FieldString {
	return &FieldString{
		Charset: charset,
		Version: version,
	}
}

func NewFieldString(charset string, regularHeader, simplifiedHeader int, stringBytes []byte, version int) *FieldString {
	fs := &FieldString{
		Charset:           charset,
		Version:           version,
		RegularOffset:     regularHeader & 0x0000FFFF,
		RegularFlags:      (regularHeader & 0x00FF0000) >> 16,
		RegularChoices:    (regularHeader & 0xFF000000) >> 24,
		SimplifiedOffset:  simplifiedHeader & 0x0000FFFF,
		SimplifiedFlags:   (simplifiedHeader & 0x00FF0000) >> 16,
		SimplifiedChoices: (simplifiedHeader & 0xFF000000) >> 24,
	}

	fs.RegularBytes = converter.GetStringBytesAtLookupOffset(stringBytes, fs.RegularOffset)

	if fs.RegularOffset == fs.SimplifiedOffset {
		fs.SimplifiedBytes = fs.RegularBytes
	} else {
		fs.SimplifiedBytes = converter.GetStringBytesAtLookupOffset(stringBytes, fs.SimplifiedOffset)
	}

	return fs
}

func FromFieldStringData(bytes []byte, charset string, version int) ([]*FieldString, error) {
	if len(bytes) == 0 {
		return []*FieldString{}, nil
	}

	first := int(bytes[0x00]) + int(bytes[0x01])*0x100
	count := first / 0x08

	strings := make([]*FieldString, 0, count)

	for i := 0; i < count; i++ {
		regularHeader := Read4Bytes(bytes, i*0x08)
		simplifiedHeader := Read4Bytes(bytes, i*0x08+0x04)
		fieldString := NewFieldString(charset, regularHeader, simplifiedHeader, bytes, version)
		strings = append(strings, fieldString)
	}

	return strings, nil
}

func RebuildFieldStrings(strings []*FieldString, charset string, version models.GameVersion) []byte {
	count := len(strings)
	contentOffset := count * 8
	offsetMap := make(map[string]int)
	var buf bytes.Buffer

	for _, fieldString := range strings {
		regularString := fieldString.GetRegularString()
		fieldString.RegularChoices = converter.GetChoicesInString(regularString)

		if regularString == "" {
			fieldString.RegularOffset = contentOffset
		} else if offset, exists := offsetMap[regularString]; exists {
			fieldString.RegularOffset = contentOffset + offset
		} else {
			fieldString.RegularOffset = contentOffset + buf.Len()
			offsetMap[regularString] = buf.Len()
			converter.FillByteList(regularString, &buf, charset, version)
		}

		simplifiedString := fieldString.GetSimplifiedString()
		fieldString.SimplifiedChoices = converter.GetChoicesInString(simplifiedString)

		if simplifiedString == "" {
			fieldString.SimplifiedOffset = contentOffset
		} else if offset, exists := offsetMap[simplifiedString]; exists {
			fieldString.SimplifiedOffset = contentOffset + offset
		} else {
			fieldString.SimplifiedOffset = contentOffset + buf.Len()
			offsetMap[simplifiedString] = buf.Len()
			converter.FillByteList(simplifiedString, &buf, charset, version)
		}
	}

	return buf.Bytes()
}

func (fs *FieldString) ToRegularHeaderBytes() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(
		fs.RegularOffset|fs.RegularFlags<<16|fs.RegularChoices<<24))
	return buf
}

func (fs *FieldString) ToSimplifiedHeaderBytes() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(
		fs.SimplifiedOffset|fs.SimplifiedFlags<<16|fs.SimplifiedChoices<<24))
	return buf
}

func (fs *FieldString) String() string {
	if fs.HasDistinctSimplified() {
		return fs.GetRegularString() + " (Simplified: " + fs.GetSimplifiedString() + ")"
	}
	return fs.GetRegularString()
}

func (fs *FieldString) IsEmpty() bool {
	return fs.GetRegularString() == "" && fs.GetSimplifiedString() == ""
}

func (fs *FieldString) GetRegularString() string {
	return converter.BytesToString(fs.RegularBytes, fs.Charset, fs.Version)
}

func (fs *FieldString) GetSimplifiedString() string {
	return converter.BytesToString(fs.SimplifiedBytes, fs.Charset, fs.Version)
}

func (fs *FieldString) HasDistinctSimplified() bool {
	if len(fs.RegularBytes) != len(fs.SimplifiedBytes) {
		return true
	}
	for i, b := range fs.RegularBytes {
		if b != fs.SimplifiedBytes[i] {
			return true
		}
	}
	return false
}

func (fs *FieldString) SetRegularString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		fs.SetCharset(newCharset[0])
	}

	keepSimplifiedSynced := !fs.HasDistinctSimplified()
	fs.RegularBytes = converter.StringToBytes(str, fs.Charset, models.GameVersion(fs.Version))

	if keepSimplifiedSynced {
		fs.SimplifiedBytes = fs.RegularBytes
	}
}

func (fs *FieldString) SetSimplifiedString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		fs.SetCharset(newCharset[0])
	}

	fs.SimplifiedBytes = converter.StringToBytes(str, fs.Charset, models.GameVersion(fs.Version))
}

func (fs *FieldString) SetCharset(newCharset string) {
	if newCharset != "" && newCharset != fs.Charset {
		fs.Charset = newCharset
	}
}

func Read4Bytes(bytes []byte, offset int) int {
	if offset+3 >= len(bytes) {
		return 0
	}
	return int(components.Read4BytesLE(bytes, offset))
}
