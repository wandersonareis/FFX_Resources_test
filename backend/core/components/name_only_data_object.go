package components

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"fmt"
)

type (
	NameOnlyDataObject struct {
		Bytes            []byte
		Name             *LocalizedKeyedStringObject
		FirstSeparator   *LocalizedKeyedStringObject
		headerParameters []byte
		HeaderLength     int
	}
	nameOnlyHeaderData struct {
		NameOffset           uint16
		NameKey              uint16
		FirstSeparatorOffset uint16
		FirstSeparatorKey    uint16
	}
)

const NameOnlyDataObjectLength int = 0x10

func getValidHeader(data []byte, max int) []byte {
	const minRequired = 0x8

	if len(data) < minRequired {
		return nil
	}

	end := max
	if end > len(data) {
		end = len(data)
	}

	return data[:end:end]
}

func NewNameOnlyDataObject(data []byte, stringBytes []byte, headerLength int, languageCode string) *NameOnlyDataObject {
	if len(data) < 8 {
		common.LogVerbose("Insufficient data to create NameOnlyDataObject!")
		return nil
	}

	n := &NameOnlyDataObject{
		Bytes:          data,
		Name:           NewLocalizedKeyedStringObject(),
		FirstSeparator: NewLocalizedKeyedStringObject(),
		HeaderLength:   headerLength,
	}
	if headerLength > NameOnlyDataObjectLength {
		n.headerParameters = data[NameOnlyDataObjectLength:headerLength]
	}
	n.mapBytes(stringBytes, languageCode)
	return n
}

func (n *NameOnlyDataObject) mapBytes(stringBytes []byte, languageCode string) {
	var hd nameOnlyHeaderData

	r := bytes.NewReader(getValidHeader(n.Bytes, NameOnlyDataObjectLength))
	if err := binary.Read(r, binary.LittleEndian, &hd); err != nil {
		fmt.Printf("Error reading NameOnlyDataObject: %v\n", err)
		return
	}
	n.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.NameOffset, hd.NameKey)
	n.FirstSeparator.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.FirstSeparatorOffset, hd.FirstSeparatorKey)
}

func (n *NameOnlyDataObject) ToBytes(languageCode string) []byte {
	result := make([]byte, n.HeaderLength)

	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.Name.GetLocalizedContent(languageCode)))
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.FirstSeparator.GetLocalizedContent(languageCode)))

	copy(result, buf.Bytes())

	if len(n.headerParameters) > 0 && NameOnlyDataObjectLength+len(n.headerParameters) <= n.HeaderLength {
		copy(result[NameOnlyDataObjectLength:], n.headerParameters)
	}

	return result
}

func (n *NameOnlyDataObject) ToList(filename string, languageCode string) IList[*NameOnlyDataObject] {
	creator := func(data []byte, stringBytes []byte, headerLength int, loc string) *NameOnlyDataObject {
		return NewNameOnlyDataObject(data, stringBytes, headerLength, loc)
	}
	return ReadDataList(filename, languageCode, creator)
}

func (n *NameOnlyDataObject) GetTextObject() ILocalizedTextObject {
	return n
}

func (n *NameOnlyDataObject) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *NameOnlyDataObject) GetKeyedString(title string) *LocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	default:
		return nil
	}
}

func (n *NameOnlyDataObject) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameOnlyDataObject) SetLocalizations(other LocalizationSetter) {
	if otherName, ok := other.(*NameOnlyDataObject); ok {
		otherName.Name.CopyInto(n.Name)
		otherName.FirstSeparator.CopyInto(n.FirstSeparator)
	}
}

func (n *NameOnlyDataObject) GetLocalizedKeyedStrings(languageCode string) []*KeyedString {
	return []*KeyedString{
		n.Name.GetLocalizedContent(languageCode),
		n.FirstSeparator.GetLocalizedContent(languageCode),
	}
}
func (d *NameOnlyDataObject) ToString(languageCode string) string {
	nameStr := d.GetName(languageCode)
	firstSepStr := ""
	if firstSepContent := d.FirstSeparator.GetLocalizedContent(languageCode); firstSepContent != nil {
		firstSepStr = firstSepContent.GetString()
	}

	return fmt.Sprintf("%s %s", nameStr, firstSepStr)
}

func (n *NameOnlyDataObject) String() string {
	return n.ToString(common.DefaultLocalization)
}
