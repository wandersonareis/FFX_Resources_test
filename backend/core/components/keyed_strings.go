package components

import "sort"

// KeyedString represents a string with offset and key for lookup tables.
type KeyedString struct {
	Charset string
	Offset  int
	Key     int
	Bytes   []byte
}

// NewKeyedString constructs a KeyedString by extracting bytes at offset
func NewKeyedString(charset string, offset, key int, data []byte) *KeyedString {
	ks := &KeyedString{
		Charset: charset,
		Offset:  offset,
		Key:     key,
	}
	ks.Bytes = GetStringBytesAtLookupOffset(data, offset)
	return ks
}

// ToHeaderBytes packs offset and key into a single uint32 (offset | key<<16)
func (ks *KeyedString) ToHeaderBytes() uint32 {
	return uint32(ks.Offset) | uint32(ks.Key)<<16
}

func (ks *KeyedString) String() string {
	return ks.GetString()
}

// GetString decodes the bytes using the charset
func (ks *KeyedString) GetString() string {
	return BytesToString(ks.Bytes, ks.Charset)
}

// IsEmpty returns true if the string is empty
func (ks *KeyedString) IsEmpty() bool {
	return ks.GetString() == ""
}

// SetString updates the string and optionally the charset
func (ks *KeyedString) SetString(str, newCharset string) {
	if newCharset != "" && newCharset != ks.Charset {
		ks.Charset = newCharset
	}
	ks.Bytes = StringToBytes(str, ks.Charset)
}

// RebuildKeyedStrings rebuilds a lookup table of strings, returning concatenated bytes
// optimize flag is reserved for future use
func RebuildKeyedStrings(strings []*KeyedString, charset string, optimize bool) []byte {
	// sort by length
	sort.Slice(strings, func(i, j int) bool {
		return len(strings[i].Bytes) < len(strings[j].Bytes)
	})
	lookup := make(map[string]*KeyedString)
	byteList := []byte{0x00}

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
		ks.Offset = len(byteList)
		ks.Key = len(lookup)
		lookup[s] = ks
		// append bytes for s
		byteList = append(byteList, StringToBytes(s, charset)...) 
	}

	return byteList
}
