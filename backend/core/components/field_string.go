package components

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// FieldString represents a field string with regular and simplified versions
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

/* func NewFieldString(charset string, regularHeader, simplifiedHeader int, data []byte) *FieldString {
	var regOffset uint16
	var regFlags uint8
	var regChoices uint8
	var simpOffset uint16
	var simpFlags uint8
	var simpChoices uint8

	r := bytes.NewReader(data)

	if err := binary.Read(r, binary.LittleEndian, &regOffset); err != nil {
		fmt.Printf("Error reading regular offset: %v\n", err)
		return nil
	}
	if err := binary.Read(r, binary.LittleEndian, &regFlags); err != nil {
		fmt.Printf("Error reading regular flags: %v\n", err)
		return nil
	}
	if err := binary.Read(r, binary.LittleEndian, &regChoices); err != nil {
		fmt.Printf("Error reading regular choices: %v\n", err)
		return nil
	}
	if err := binary.Read(r, binary.LittleEndian, &simpOffset); err != nil {
		fmt.Printf("Error reading simplified offset: %v\n", err)
		return nil
	}
	if err := binary.Read(r, binary.LittleEndian, &simpFlags); err != nil {
		fmt.Printf("Error reading simplified flags: %v\n", err)
		return nil
	}
	if err := binary.Read(r, binary.LittleEndian, &simpChoices); err != nil {
		fmt.Printf("Error reading simplified choices: %v\n", err)
		return nil
	}
	fs := &FieldString{
		Charset:           charset,
		RegularOffset:     regOffset,
		RegularFlags:      regFlags,
		RegularChoices:    regChoices,
		SimplifiedOffset:  simpOffset,
		SimplifiedFlags:   simpFlags,
		SimplifiedChoices: simpChoices,
	}

	fs.RegularBytes = GetStringBytesAtLookupOffset(data, fs.RegularOffset)

	if fs.RegularOffset == fs.SimplifiedOffset {
		fs.SimplifiedBytes = fs.RegularBytes
	} else {
		fs.SimplifiedBytes = GetStringBytesAtLookupOffset(data, fs.SimplifiedOffset)
	}

	return fs
} */

// NewFieldString creates a new FieldString from header data and bytes
func NewFieldString(charset string, regularHeader, simplifiedHeader int, bytes []byte) *FieldString {
	fs := &FieldString{
		Charset:           charset,
		RegularOffset:     regularHeader & 0x0000FFFF,
		RegularFlags:      (regularHeader & 0x00FF0000) >> 16,
		RegularChoices:    (regularHeader & 0xFF000000) >> 24,
		SimplifiedOffset:  simplifiedHeader & 0x0000FFFF,
		SimplifiedFlags:   (simplifiedHeader & 0x00FF0000) >> 16,
		SimplifiedChoices: (simplifiedHeader & 0xFF000000) >> 24,
	}

	fs.RegularBytes = GetStringBytesAtLookupOffset(bytes, fs.RegularOffset)

	if fs.RegularOffset == fs.SimplifiedOffset {
		fs.SimplifiedBytes = fs.RegularBytes
	} else {
		fs.SimplifiedBytes = GetStringBytesAtLookupOffset(bytes, fs.SimplifiedOffset)
	}

	return fs
}

// FromFieldStringData creates a list of FieldString objects from byte data
func FromFieldStringData(bytes []byte, print bool, charset string) ([]*FieldString, error) {
	if len(bytes) == 0 {
		return []*FieldString{}, nil
	}

	// Read first two bytes to get count
	first := int(bytes[0x00]) + int(bytes[0x01])*0x100
	count := first / 0x08

	strings := make([]*FieldString, 0, count)

	for i := 0; i < count; i++ {
		regularHeader := Read4Bytes(bytes, i*0x08)
		simplifiedHeader := Read4Bytes(bytes, i*0x08+0x04)

		fieldString := NewFieldString(charset, regularHeader, simplifiedHeader, bytes)

		if print {
			fmt.Printf("String %02X: %s\n", i, fieldString.String())
		}

		strings = append(strings, fieldString)
	}

	return strings, nil
}

// RebuildFieldStrings rebuilds field strings into byte array format
func RebuildFieldStrings(strings []*FieldString, charset string, optimize bool) []byte {
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

		// Handle simplified string
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

// ToRegularHeaderBytes converts the regular header fields to a 4-byte integer
func (fs *FieldString) ToRegularHeaderBytes() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(
		fs.RegularOffset|fs.RegularFlags<<16|fs.RegularChoices<<24))
	return buf
}

// ToSimplifiedHeaderBytes converts the simplified header fields to a 4-byte integer
func (fs *FieldString) ToSimplifiedHeaderBytes() []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(
		fs.SimplifiedOffset|fs.SimplifiedFlags<<16|fs.SimplifiedChoices<<24))
	return buf
}

// String returns the string representation of the FieldString
func (fs *FieldString) String() string {
	if fs.HasDistinctSimplified() {
		return fs.GetRegularString() + " (Simplified: " + fs.GetSimplifiedString() + ")"
	}
	return fs.GetRegularString()
}

// IsEmpty returns true if both regular and simplified strings are empty
func (fs *FieldString) IsEmpty() bool {
	return fs.GetRegularString() == "" && fs.GetSimplifiedString() == ""
}

// GetRegularString returns the regular string converted from bytes
func (fs *FieldString) GetRegularString() string {
	return BytesToString(fs.RegularBytes, fs.Charset)
}

// GetSimplifiedString returns the simplified string converted from bytes
func (fs *FieldString) GetSimplifiedString() string {
	return BytesToString(fs.SimplifiedBytes, fs.Charset)
}

// HasDistinctSimplified returns true if simplified string is different from regular string
func (fs *FieldString) HasDistinctSimplified() bool {
	// Compare byte slices
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

// SetRegularString sets the regular string with optional charset change
func (fs *FieldString) SetRegularString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		fs.SetCharset(newCharset[0])
	}

	keepSimplifiedSynced := !fs.HasDistinctSimplified()
	fs.RegularBytes = StringToBytes(str, fs.Charset)
	fmt.Println("Setting regular string:", fs.GetRegularString(), "with charset:", fs.Charset)

	if keepSimplifiedSynced {
		fs.SimplifiedBytes = fs.RegularBytes
	}
}

// SetSimplifiedString sets the simplified string with optional charset change
func (fs *FieldString) SetSimplifiedString(str string, newCharset ...string) {
	if len(newCharset) > 0 && newCharset[0] != "" {
		fs.SetCharset(newCharset[0])
	}

	fs.SimplifiedBytes = StringToBytes(str, fs.Charset)
}

// SetCharset updates the charset if different from current
func (fs *FieldString) SetCharset(newCharset string) {
	if newCharset != "" && newCharset != fs.Charset {
		fs.Charset = newCharset
	}
}

// Helper functions that need to be implemented or already exist in the project

// Read4Bytes reads 4 bytes from the byte array at the given offset as little-endian integer
func Read4Bytes(bytes []byte, offset int) int {
	if offset+3 >= len(bytes) {
		return 0
	}
	return int(read4BytesLE(bytes, offset))
}

// Hex2WithSuffix formats an integer as a 2-digit hex string with suffix
func Hex2WithSuffix(value int) string {
	return fmt.Sprintf("%02X", value)
}
