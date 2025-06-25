package components

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"fmt"
)

type (
	ILocalizedTextObject interface {
		GetName(languageCode string) string
		GetKeyedString(title string) *LocalizedKeyedStringObject
		GetLocalizedKeyedStrings(languageCode string) []*KeyedString
		SetLocalizations(other LocalizationSetter)
		GetTextObject() ILocalizedTextObject
		GetHeaderLength() int
		ToBytes(languageCode string) []byte
		ToString(languageCode string) string
		String() string
	}
	NameOnlyTextObject struct {
		Bytes          []byte
		Name           *LocalizedKeyedStringObject
		FirstSeparator *LocalizedKeyedStringObject
		HeaderLength   int
	}
	nameOnlyHeaderData struct {
		NameOffset           uint16
		NameKey              uint16
		FirstSeparatorOffset uint16
		FirstSeparatorKey    uint16
	}
)

const NameOnlyTextObjectLength = 0x10

func getLocalizedBytes(keyedObject *KeyedString) []uint16 {
	if keyedObject == nil {
		return []uint16{0, 0}
	}
	return []uint16{keyedObject.Offset, keyedObject.Key}
}

func NewNameOnlyTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string) *NameOnlyTextObject {
	n := &NameOnlyTextObject{
		Bytes:          bytes,
		Name:           NewLocalizedKeyedStringObject(),
		FirstSeparator: NewLocalizedKeyedStringObject(),
		HeaderLength:   headerLength,
	}
	n.mapBytes(stringBytes, languageCode)
	return n
}

func (n *NameOnlyTextObject) mapBytes(stringBytes []byte, languageCode string) {
	var hd nameOnlyHeaderData

	r := bytes.NewReader(n.Bytes[:NameOnlyTextObjectLength])
	if err := binary.Read(r, binary.LittleEndian, &hd); err != nil {
		fmt.Printf("Error reading NameOnlyTextObject: %v\n", err)
		return
	}
	n.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.NameOffset, hd.NameKey)
	n.FirstSeparator.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.FirstSeparatorOffset, hd.FirstSeparatorKey)
}

func (n *NameOnlyTextObject) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (d *NameOnlyTextObject) GetKeyedString(title string) *LocalizedKeyedStringObject {
	switch title {
	case "name":
		return d.Name
	default:
		return nil
	}
}

func (n *NameOnlyTextObject) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameOnlyTextObject) ToBytes(languageCode string) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.Name.GetLocalizedContent(languageCode)))
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.FirstSeparator.GetLocalizedContent(languageCode)))

	return buf.Bytes()
}

func (n *NameOnlyTextObject) GetTextObject() ILocalizedTextObject {
	return n
}

func (n *NameOnlyTextObject) SetLocalizations(other LocalizationSetter) {
	if otherName, ok := other.(*NameOnlyTextObject); ok {
		otherName.Name.CopyInto(n.Name)
		otherName.FirstSeparator.CopyInto(n.FirstSeparator)
	}
}

func (n *NameOnlyTextObject) GetLocalizedKeyedStrings(languageCode string) []*KeyedString {
	return []*KeyedString{
		n.Name.GetLocalizedContent(languageCode),
		n.FirstSeparator.GetLocalizedContent(languageCode),
	}
}

func (d *NameOnlyTextObject) ToString(languageCode string) string {
	nameStr := d.GetName(languageCode)
	firstSepStr := ""
	if firstSepContent := d.FirstSeparator.GetLocalizedContent(languageCode); firstSepContent != nil {
		firstSepStr = firstSepContent.GetString()
	}

	return fmt.Sprintf("%s %s", nameStr, firstSepStr)
}

func (n *NameOnlyTextObject) String() string {
	return n.ToString(common.DefaultLocalization)
}

type (
	NameDescriptionTextObject struct {
		Bytes            []byte
		Name             *LocalizedKeyedStringObject
		FirstSeparator   *LocalizedKeyedStringObject
		Description      *LocalizedKeyedStringObject
		SecondSeparator  *LocalizedKeyedStringObject
		headerParameters []byte
		HeaderLength     int
	}
	nameDescriptionHeaderData struct {
		NameOffset            uint16
		NameKey               uint16
		FirstSeparatorOffset  uint16
		FirstSeparatorKey     uint16
		DescriptionOffset     uint16
		DescriptionKey        uint16
		SecondSeparatorOffset uint16
		SecondSeparatorKey    uint16
	}
)

const NameDescriptionTextObjectLength = 0x10

func NewNameDescriptionTextObject(bytes []byte, stringBytes []byte, headerLength int, localization string) *NameDescriptionTextObject {
	n := &NameDescriptionTextObject{
		Bytes:           bytes,
		Name:            NewLocalizedKeyedStringObject(),
		FirstSeparator:  NewLocalizedKeyedStringObject(),
		Description:     NewLocalizedKeyedStringObject(),
		SecondSeparator: NewLocalizedKeyedStringObject(),
		HeaderLength:    headerLength,
	}
	n.mapBytes(stringBytes, localization)
	return n
}

func (n *NameDescriptionTextObject) mapBytes(stringBytes []byte, localization string) {
	var hd nameDescriptionHeaderData

	r := bytes.NewReader(n.Bytes[:NameDescriptionTextObjectLength])
	if err := binary.Read(r, binary.LittleEndian, &hd); err != nil {
		fmt.Printf("Error reading NameDescriptionTextObject: %v\n", err)
		return
	}

	n.Name.ReadAndSetLocalizedContent(localization, stringBytes, hd.NameOffset, hd.NameKey)
	n.FirstSeparator.ReadAndSetLocalizedContent(localization, stringBytes, hd.FirstSeparatorOffset, hd.FirstSeparatorKey)
	n.Description.ReadAndSetLocalizedContent(localization, stringBytes, hd.DescriptionOffset, hd.DescriptionKey)
	n.SecondSeparator.ReadAndSetLocalizedContent(localization, stringBytes, hd.SecondSeparatorOffset, hd.SecondSeparatorKey)

	if n.HeaderLength > NameDescriptionTextObjectLength {
		n.headerParameters = n.Bytes[NameDescriptionTextObjectLength:n.HeaderLength]
	}
}

func (n *NameDescriptionTextObject) ToBytes(localization string) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.Name.GetLocalizedContent(localization)))
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.FirstSeparator.GetLocalizedContent(localization)))
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.Description.GetLocalizedContent(localization)))
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.SecondSeparator.GetLocalizedContent(localization)))

	if len(n.headerParameters) > 0 && NameDescriptionTextObjectLength+len(n.headerParameters) <= n.HeaderLength {
		buf.Write(n.headerParameters)
	}
	return buf.Bytes()
}

func (n *NameDescriptionTextObject) GetName(localization string) string {
	return n.Name.GetLocalizedString(localization)
}

func (d *NameDescriptionTextObject) GetKeyedString(title string) *LocalizedKeyedStringObject {
	switch title {
	case "name":
		return d.Name
	case "description":
		return d.Description
	default:
		return nil
	}
}

func (n *NameDescriptionTextObject) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameDescriptionTextObject) GetTextObject() ILocalizedTextObject {
	return n
}

func (n *NameDescriptionTextObject) SetLocalizations(other LocalizationSetter) {
	if otherNameDesc, ok := other.(*NameDescriptionTextObject); ok {
		otherNameDesc.Name.CopyInto(n.Name)
		otherNameDesc.FirstSeparator.CopyInto(n.FirstSeparator)
		otherNameDesc.Description.CopyInto(n.Description)
		otherNameDesc.SecondSeparator.CopyInto(n.SecondSeparator)
	}
}

func (n *NameDescriptionTextObject) GetLocalizedKeyedStrings(localization string) []*KeyedString {
	return []*KeyedString{
		n.Name.GetLocalizedContent(localization),
		n.FirstSeparator.GetLocalizedContent(localization),
		n.Description.GetLocalizedContent(localization),
		n.SecondSeparator.GetLocalizedContent(localization),
	}
}

func (d *NameDescriptionTextObject) ToString(languageCode string) string {
	nameStr := d.GetName(languageCode)
	firstSepStr := ""
	if firstSepContent := d.FirstSeparator.GetLocalizedContent(languageCode); firstSepContent != nil {
		firstSepStr = firstSepContent.GetString()
	}

	descStr := ""
	if descContent := d.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	secondSepStr := ""
	if secondSepContent := d.SecondSeparator.GetLocalizedContent(languageCode); secondSepContent != nil {
		secondSepStr = secondSepContent.GetString()
	}

	return fmt.Sprintf("%s %s %s %s", nameStr, firstSepStr, descStr, secondSepStr)
}

func (n *NameDescriptionTextObject) String() string {
	return n.ToString(common.DefaultLocalization)
}
