package objectsfile

import (
	"bytes"
	"fmt"
	"io"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

type NameDescriptionTextObjectV2 struct {
	Bytes        []byte
	Name         datastore.IGlobalLocalizedKeyedStringObject
	Description  datastore.IGlobalLocalizedKeyedStringObject
	unknownBytes []byte
	HeaderLength int
}

const NameDescriptionTextObjectV2Length = 0x08

func NewNameDescriptionTextObjectV2(bytes []byte, stringBytes []byte, headerLength int, languageCode string) (*NameDescriptionTextObjectV2, error) {
	if len(bytes) < NameDescriptionTextObjectV2Length {
		return nil, fmt.Errorf("insufficient data: have %d bytes, need at least %d", len(bytes), NameDescriptionTextObjectV2Length)
	}

	n := &NameDescriptionTextObjectV2{
		Bytes:        bytes,
		Name:         NewLocalizedKeyedStringObject(),
		Description:  NewLocalizedKeyedStringObject(),
		HeaderLength: headerLength,
	}

	if err := n.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *NameDescriptionTextObjectV2) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(n.Bytes)

	if err := readStringSegments(r, stringBytes, languageCode, n.Name, n.Description); err != nil {
		common.LogError("Error reading NameDescriptionTextObjectV2 segments: %v", err)
		return err
	}

	if r.Len() > 0 {
		n.unknownBytes = make([]byte, r.Len())
		if _, err := io.ReadFull(r, n.unknownBytes); err != nil {
			common.LogError("Error reading NameDescriptionTextObjectV2 unknown bytes: %v", err)
			return err
		}
	}
	return nil
}

func (n *NameDescriptionTextObjectV2) ToBytes(languageCode string) ([]byte, error) {
	segLen := NameDescriptionTextObjectV2Length
	if len(n.unknownBytes) > 0 {
		segLen += len(n.unknownBytes)
	}
	result := make([]byte, segLen)

	if err := writeStringSegments(result, 0, languageCode,
		n.Name,
		n.Description,
	); err != nil {
		return nil, err
	}

	if len(n.unknownBytes) > 0 {
		copy(result[NameDescriptionTextObjectV2Length:], n.unknownBytes)
	}
	return result, nil
}

func (n *NameDescriptionTextObjectV2) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *NameDescriptionTextObjectV2) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	case "description":
		return n.Description
	default:
		return nil
	}
}

func (n *NameDescriptionTextObjectV2) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameDescriptionTextObjectV2) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameDescriptionTextObjectV2) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameDescriptionTextObjectV2); ok {
		o.Name.CopyInto(n.Name)
		o.Description.CopyInto(n.Description)
		if len(o.unknownBytes) > 0 {
			n.unknownBytes = append([]byte{}, o.unknownBytes...)
		}
	}
}

func (n *NameDescriptionTextObjectV2) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Name.GetLocalizedContent(localization),
		n.Description.GetLocalizedContent(localization),
	}
}

func (n *NameDescriptionTextObjectV2) ToString(languageCode string) string {
	nameStr := n.GetName(languageCode)
	descStr := ""
	if descContent := n.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	return fmt.Sprintf("%s - %s", nameStr, descStr)
}

func (n *NameDescriptionTextObjectV2) String() string {
	return n.ToString(common.DefaultLocalization)
}
