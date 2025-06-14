package components

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type (
	headerData struct {
		NameOffset            uint16
		NameKey               uint16
		FirstSeparatorOffset  uint16
		FirstSeparatorKey     uint16
		DescriptionOffset     uint16
		DescriptionKey        uint16
		SecondSeparatorOffset uint16
		SecondSeparatorKey    uint16
	}
	NameDescriptionTextObject struct {
		headerData      headerData
		Bytes           []byte
		Name            *LocalizedKeyedStringObject
		FirstSeparator  *LocalizedKeyedStringObject
		Description     *LocalizedKeyedStringObject
		SecondSeparator *LocalizedKeyedStringObject
	}
)

const NameDescriptionTextObjectLength = 0x10

func NewNameDescriptionTextObject(bytes []byte, stringBytes []byte, localization string) *NameDescriptionTextObject {
	n := &NameDescriptionTextObject{
		Bytes:           bytes,
		Name:            NewLocalizedKeyedStringObject(),
		FirstSeparator:  NewLocalizedKeyedStringObject(),
		Description:     NewLocalizedKeyedStringObject(),
		SecondSeparator: NewLocalizedKeyedStringObject(),
	}
	n.mapBytes()
	n.mapStrings(stringBytes, localization)
	return n
}

func (n *NameDescriptionTextObject) mapBytes() {
	var header headerData
	r := bytes.NewReader(n.Bytes)

	if err := binary.Read(r, binary.LittleEndian, &header); err != nil {
		fmt.Printf("Error reading NameDescriptionTextObject: %v\n", err)
		return
	}
	n.headerData = header
}

func (n *NameDescriptionTextObject) mapStrings(table []byte, localization string) {
	n.Name.ReadAndSetLocalizedContent(localization, table, n.headerData.NameOffset, n.headerData.NameKey)
	n.FirstSeparator.ReadAndSetLocalizedContent(localization, table, n.headerData.FirstSeparatorOffset, n.headerData.FirstSeparatorKey)
	n.Description.ReadAndSetLocalizedContent(localization, table, n.headerData.DescriptionOffset, n.headerData.DescriptionKey)
	n.SecondSeparator.ReadAndSetLocalizedContent(localization, table, n.headerData.SecondSeparatorOffset, n.headerData.SecondSeparatorKey)
}

func (n *NameDescriptionTextObject) ToBytes(localization string) []byte {
	buf := new(bytes.Buffer)
	n.Name.GetLocalizedContent(localization).GetHeaderBytes(buf)
	n.FirstSeparator.GetLocalizedContent(localization).GetHeaderBytes(buf)
	n.Description.GetLocalizedContent(localization).GetHeaderBytes(buf)
	n.SecondSeparator.GetLocalizedContent(localization).GetHeaderBytes(buf)
	return buf.Bytes()
}

func (n *NameDescriptionTextObject) GetName(localization string) string {
	return n.Name.GetLocalizedString(localization)
}

func (n *NameDescriptionTextObject) SetLocalizations(other *NameDescriptionTextObject) {
	other.Name.CopyInto(n.Name)
	other.FirstSeparator.CopyInto(n.FirstSeparator)
	other.Description.CopyInto(n.Description)
	other.SecondSeparator.CopyInto(n.SecondSeparator)
}

func (n *NameDescriptionTextObject) GetLocalizedKeyedStrings(localization string) []*KeyedString {
	return []*KeyedString{
		n.Name.GetLocalizedContent(localization),
		n.FirstSeparator.GetLocalizedContent(localization),
		n.Description.GetLocalizedContent(localization),
		n.SecondSeparator.GetLocalizedContent(localization),
	}
}

func (n *NameDescriptionTextObject) String() string {
	descStr := ""
	if n.headerData.DescriptionOffset > 0 {
		descStr = n.Description.GetDefaultContent().String()
	}
	return fmt.Sprintf("%-20s - %s", n.GetName(""), descStr)
}
