package components

import (
	"bytes"
	"encoding/binary"
)

type FieldString struct {
	Charset           string
	RegularOffset     int
	RegularFlags      int
	RegularChoices    int
	SimplifiedOffset  int
	SimplifiedFlags   int
	SimplifiedChoices int
	RegularBytes      []byte
	SimplifiedBytes   []byte
}

func NewFieldString(charset string, regularHeader, simplifiedHeader int, stringBytes []byte) *FieldString {
	fs := &FieldString{
		Charset:           charset,
		RegularOffset:     regularHeader & 0x0000FFFF,
		RegularFlags:      (regularHeader & 0x00FF0000) >> 16,
		RegularChoices:    (regularHeader & 0xFF000000) >> 24,
		SimplifiedOffset:  simplifiedHeader & 0x0000FFFF,
		SimplifiedFlags:   (simplifiedHeader & 0x00FF0000) >> 16,
		SimplifiedChoices: (simplifiedHeader & 0xFF000000) >> 24,
	}

	fs.RegularBytes = GetStringBytesAtLookupOffset(stringBytes, fs.RegularOffset)

	if fs.RegularOffset == fs.SimplifiedOffset {
		fs.SimplifiedBytes = fs.RegularBytes
	} else {
		fs.SimplifiedBytes = GetStringBytesAtLookupOffset(stringBytes, fs.SimplifiedOffset)
	}

	return fs
}

func FromFieldStringData(bytes []byte, charset string) ([]*FieldString, error) {
	if len(bytes) == 0 {
		return []*FieldString{}, nil
	}

	first := int(bytes[0x00]) + int(bytes[0x01])*0x100
	count := first / 0x08

	strings := make([]*FieldString, 0, count)

	for i := 0; i < count; i++ {
		regularHeader := Read4Bytes(bytes, i*0x08)
		simplifiedHeader := Read4Bytes(bytes, i*0x08+0x04)
		fieldString := NewFieldString(charset, regularHeader, simplifiedHeader, bytes)
		strings = append(strings, fieldString)
	}

	return strings, nil
}

func RebuildFieldStrings(strings []*FieldString, charset string) []byte {
	count := len(strings)
	contentOffset := count * 8
	offsetMap := make(map[string]int)
	var buf bytes.Buffer

	for _, fieldString := range strings {
		regularString := fieldString.GetRegularString()
		fieldString.RegularChoices = GetChoicesInString(regularString)

		if regularString == "" {
			fieldString.RegularOffset = contentOffset
		} else if offset, exists := offsetMap[regularString]; exists {
			fieldString.RegularOffset = contentOffset + offset
		} else {
			fieldString.RegularOffset = contentOffset + buf.Len()
			offsetMap[regularString] = buf.Len()
			FillByteList(regularString, &buf, charset)
		}

		simplifiedString := fieldString.GetSimplifiedString()
		fieldString.SimplifiedChoices = GetChoicesInString(simplifiedString)

		if simplifiedString == "" {
			fieldString.SimplifiedOffset = contentOffset
		} else if offset, exists := offsetMap[simplifiedString]; exists {
			fieldString.SimplifiedOffset = contentOffset + offset
		} else {
			fieldString.SimplifiedOffset = contentOffset + buf.Len()
			offsetMap[simplifiedString] = buf.Len()
			FillByteList(simplifiedString, &buf, charset)
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
	return BytesToString(fs.RegularBytes, fs.Charset)
}

func (fs *FieldString) GetSimplifiedString() string {
	return BytesToString(fs.SimplifiedBytes, fs.Charset)
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
	fs.RegularBytes = StringToBytes(str, fs.Charset)

	if keepSimplifiedSynced {
		fs.SimplifiedBytes = fs.RegularBytes
	}
}

func (fs *FieldString) SetSimplifiedString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		fs.SetCharset(newCharset[0])
	}

	fs.SimplifiedBytes = StringToBytes(str, fs.Charset)
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
	return int(read4BytesLE(bytes, offset))
}
