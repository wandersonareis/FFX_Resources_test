package components

import (
	"bytes"
	"encoding/binary"
)

type KeyedString struct {
	Charset string
	Offset  uint16
	Key     uint16
	Bytes   []byte
}

func NewKeyedString(charset string, offset, key uint16, data []byte) *KeyedString {
	ks := &KeyedString{
		Charset: charset,
		Offset:  offset,
		Key:     key,
	}
	ks.Bytes = GetStringBytesAtLookupOffset(data, int(offset))
	return ks
}

func (ks *KeyedString) GetHeaderBytes(buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, uint16(ks.Offset))
	binary.Write(buf, binary.LittleEndian, uint16(ks.Key))
}

func (ks *KeyedString) String() string {
	return ks.GetString()
}

func (ks *KeyedString) GetString() string {
	return BytesToString(ks.Bytes, ks.Charset)
}

func (ks *KeyedString) IsEmpty() bool {
	return ks.GetString() == ""
}

func (ks *KeyedString) SetString(str, newCharset string) {
	if newCharset != "" && newCharset != ks.Charset {
		ks.Charset = newCharset
	}
	ks.Bytes = StringToBytes(str, ks.Charset)
}

func RebuildKeyedStrings(strings []*KeyedString, charset string) []byte {
	lookup := make(map[string]*KeyedString)
	var buf bytes.Buffer

	for _, ks := range strings {
		s := ks.GetString()

		if s == "" {
			ks.Offset, ks.Key = 0, 0
			continue
		}

		if existing, ok := lookup[s]; ok {
			ks.Offset, ks.Key = existing.Offset, existing.Key
			continue
		}
		// new entry
		ks.Offset = uint16(buf.Len())
		lookup[s] = ks
		FillByteList(s, &buf, charset)
	}

	return buf.Bytes()
}
