package objectsfile

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"fmt"
	"slices"
)

type DescriptionOnlyTextObject struct {
	Bytes                 []byte
	Description           datastore.IGlobalLocalizedKeyedStringObject
	SimplifiedDescription datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength          int
}

const DescriptionOnlyTextObjectLength = 0x08

func NewDescriptionOnlyTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string, gameVersion int) (*DescriptionOnlyTextObject, error) {
	if gameVersion != 1 {
		return nil, fmt.Errorf("DescriptionOnlyTextObject is only compatible with FFX (game version 1), but got game version %d", gameVersion)
	}
	
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	n := &DescriptionOnlyTextObject{
		Bytes:                 bytes,
		Description:           NewLocalizedKeyedStringObject(),
		SimplifiedDescription: NewLocalizedKeyedStringObject(),
		HeaderLength:          headerLength,
	}

	if err := n.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *DescriptionOnlyTextObject) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(getValidHeader(n.Bytes, DescriptionOnlyTextObjectLength))
	return readStringSegments(r, stringBytes, languageCode,
		n.Description,
		n.SimplifiedDescription,
	)
}

func (n *DescriptionOnlyTextObject) ToBytes(languageCode string) ([]byte, error) {
	result := slices.Clone(n.Bytes)

	if err := writeStringSegments(result, 0, languageCode,
		n.Description,
		n.SimplifiedDescription,
	); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *DescriptionOnlyTextObject) GetName(languageCode string) string {
	return ""
}

func (d *DescriptionOnlyTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "description":
		return d.Description
	case "simplifiedDescription":
		return d.SimplifiedDescription
	default:
		return nil
	}
}

func (n *DescriptionOnlyTextObject) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *DescriptionOnlyTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *DescriptionOnlyTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*DescriptionOnlyTextObject); ok {
		o.Description.CopyInto(n.Description)
		o.SimplifiedDescription.CopyInto(n.SimplifiedDescription)
	}
}

func (n *DescriptionOnlyTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Description.GetLocalizedContent(localization),
		n.SimplifiedDescription.GetLocalizedContent(localization),
	}
}

func (d *DescriptionOnlyTextObject) ToString(languageCode string) string {
	descStr := ""
	if descContent := d.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	return descStr
}

func (n *DescriptionOnlyTextObject) String() string {
	return n.ToString(common.DefaultLocalization)
}
