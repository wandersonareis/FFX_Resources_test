package objectsfile

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type KeyedString struct {
	Charset string
	Version int
/* 	Offset  uint16
	Key     uint16 */
	Segment models.Segment
	Bytes   []byte
}

func NewKeyedString(charset string, segment models.Segment, data []byte, version int) datastore.IGlobalKeyedString {
	if segment.Offset == 0 && segment.Key == 0 {
		return nil
	}

	ks := &KeyedString{
		Charset: charset,
		Version: version,
		/* Offset:  segment.Offset,
		Key:     segment.Key, */
		Segment: segment,
	}
	ks.Bytes = converter.GetStringBytesAtLookupOffset(data, int(segment.Offset))
	return ks
}

func (ks *KeyedString) GetOffset() models.Offset {
	return ks.Segment.Offset
}

func (ks *KeyedString) SetOffset(offset models.Offset) {
	if offset != ks.Segment.Offset {
		ks.Segment.Offset = offset
	}
}

func (ks *KeyedString) GetKey() models.Key {
	return ks.Segment.Key
}

func (ks *KeyedString) SetKey(key models.Key) {
	if key != ks.Segment.Key {
		ks.Segment.Key = key
	}
}

func (ks *KeyedString) GetHeaderBytes(buf *bytes.Buffer) {
	binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Offset))
	binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Key))
}

func (ks *KeyedString) SetHeaderBytes(buf *bytes.Buffer) {
	if buf == nil {
		return
	}
	if len(ks.Bytes) == 0 {
		ks.Bytes = make([]byte, 0)
	}
	if ks.Segment.Offset == 0 && ks.Segment.Key == 0 {
		return
	}
	/* binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Offset))
	binary.Write(buf, binary.LittleEndian, uint16(ks.Segment.Key)) */
	if err := models.WriteSegment(buf, ks.Segment); err != nil {
		return
	}
}

func (ks *KeyedString) GetCharset() string {
	if ks.Charset == "" {
		return common.DefaultLocalization
	}
	return ks.Charset
}

func (ks *KeyedString) SetCharset(charset string) {
	if charset != "" && charset != ks.Charset {
		ks.Charset = charset
	}
}

func (ks *KeyedString) String() string {
	return ks.GetString()
}

func (ks *KeyedString) GetString() string {
	return converter.BytesToString(ks.Bytes, ks.Charset, ks.Version)
}

func (ks *KeyedString) IsEmpty() bool {
	return ks.GetString() == ""
}

func (ks *KeyedString) SetString(str, newCharset string) {
	if newCharset != "" && newCharset != ks.Charset {
		ks.Charset = newCharset
	}
	ks.Bytes = converter.StringToBytes(str, ks.Charset, models.GameVersion(ks.Version))
}

func RebuildKeyedStrings(strings []datastore.IGlobalKeyedString, charset string, version models.GameVersion) []byte {
	var buf bytes.Buffer

	for _, ks := range strings {
		s := ks.GetString()

		ks.SetOffset(models.Offset(buf.Len()))
		converter.FillByteList(s, &buf, charset, version)
	}

	return buf.Bytes()
}
