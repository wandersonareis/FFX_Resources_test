package objectsfile

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"slices"
)

type (
	NameOnlyTextObject struct {
		Bytes          []byte
		Name           datastore.IGlobalLocalizedKeyedStringObject
		SimplifiedName datastore.IGlobalLocalizedKeyedStringObject
		HeaderLength   int
	}
)

type NameOnlyTextObjectV2 struct {
	Bytes        []byte
	Name         datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength int
	unknown      []byte
}

var (
	//ARMS_TEXT   components.IList[datastore.IGlobalLocalizedTextObject]
	//BTL_TEXT    components.IList[datastore.IGlobalLocalizedTextObject]
	//BTLEND_TEXT components.IList[datastore.IGlobalLocalizedTextObject]
	BUILD_TEXT  components.IList[datastore.IGlobalLocalizedTextObject]
	CONFIG_TEXT components.IList[datastore.IGlobalLocalizedTextObject]
	ITEM_TEXT   components.IList[datastore.IGlobalLocalizedTextObject]
	MENU_TEXT   components.IList[datastore.IGlobalLocalizedTextObject]
	MMAIN_TEXT  components.IList[datastore.IGlobalLocalizedTextObject]
	NAME_TEXT   components.IList[datastore.IGlobalLocalizedTextObject]
	PLAYER_ROOM components.IList[datastore.IGlobalLocalizedTextObject]
	SAVE_TEXT   components.IList[datastore.IGlobalLocalizedTextObject]
	STATS_TEXT  components.IList[datastore.IGlobalLocalizedTextObject]
	SUMMON_TEXT components.IList[datastore.IGlobalLocalizedTextObject]
)

const NameOnlyTextObjectLength int = 0x10

func getValidHeader(data []byte, max int) []byte {
	const minRequired = 0x8

	if len(data) < minRequired {
		return nil
	}

	end := min(max, len(data))

	return data[:end:end]
}

func NewNameOnlyTextObject(data []byte, stringBytes []byte, headerLength int, languageCode string) (*NameOnlyTextObject, error) {
	if len(data) < headerLength {
		return nil, fmt.Errorf("insufficient data to create NameOnlyTextObject: have %d bytes, need at least %d", len(data), headerLength)
	}

	n := &NameOnlyTextObject{
		Bytes:          data,
		Name:           NewLocalizedKeyedStringObject(),
		SimplifiedName: NewLocalizedKeyedStringObject(),
		HeaderLength:   headerLength,
	}
	if err := n.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *NameOnlyTextObject) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(getValidHeader(n.Bytes, NameOnlyTextObjectLength))
	return readStringSegments(r, stringBytes, languageCode,
		n.Name,
		n.SimplifiedName,
	)
}

func (n *NameOnlyTextObject) ToBytes(languageCode string) ([]byte, error) {
	result := slices.Clone(n.Bytes)

	if err := writeStringSegments(result, 0, languageCode,
		n.Name,
		n.SimplifiedName,
	); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *NameOnlyTextObject) ToList(filename string, languageCode string) components.IList[datastore.IGlobalLocalizedTextObject] {
	creator := func(data []byte, stringBytes []byte, headerLength int, loc string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewNameOnlyTextObject(data, stringBytes, headerLength, loc)
	}
	return ReadDataListWithIlist(filename, languageCode, creator)
}

func (n *NameOnlyTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameOnlyTextObject) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *NameOnlyTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	case "simplifiedName":
		return n.SimplifiedName
	default:
		return nil
	}
}

func (n *NameOnlyTextObject) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameOnlyTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherName, ok := other.(*NameOnlyTextObject); ok {
		otherName.Name.CopyInto(n.Name)
		otherName.SimplifiedName.CopyInto(n.SimplifiedName)
	}
}

func (n *NameOnlyTextObject) GetLocalizedKeyedStrings(languageCode string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Name.GetLocalizedContent(languageCode),
		n.SimplifiedName.GetLocalizedContent(languageCode),
	}
}

func (d *NameOnlyTextObject) ToString(languageCode string) string {
	nameStr := d.GetName(languageCode)
	simplifiedNameStr := ""
	if simplifiedNameContent := d.SimplifiedName.GetLocalizedContent(languageCode); simplifiedNameContent != nil {
		simplifiedNameStr = simplifiedNameContent.GetString()
	}

	return fmt.Sprintf("%s %s", nameStr, simplifiedNameStr)
}

func (n *NameOnlyTextObject) String() string {
	return n.ToString(common.DefaultLocalization)
}

func NewNameOnlyTextObjectV2(data []byte, stringBytes []byte, headerLength int, languageCode string) (*NameOnlyTextObjectV2, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("insufficient data to create NameOnlyTextObjectV2: have %d bytes, need at least 4", len(data))
	}
	n := &NameOnlyTextObjectV2{
		Bytes:        data,
		Name:         NewLocalizedKeyedStringObject(),
		HeaderLength: headerLength,
	}
	if err := n.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *NameOnlyTextObjectV2) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(n.Bytes)

	if err := readStringSegments(r, stringBytes, languageCode, n.Name); err != nil {
		common.LogError("Error reading NameOnlyTextObjectV2 name segment: %v", err)
		return err
	}

	if r.Len() > 0 {
		n.unknown = make([]byte, r.Len())
		if _, err := r.Read(n.unknown); err != nil {
			common.LogError("Error reading NameOnlyTextObjectV2 unknown bytes: %v", err)
			return err
		}
	}
	return nil
}

func (n *NameOnlyTextObjectV2) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *NameOnlyTextObjectV2) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	default:
		return nil
	}
}

func (n *NameOnlyTextObjectV2) GetLocalizedKeyedStrings(languageCode string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Name.GetLocalizedContent(languageCode),
	}
}

func (n *NameOnlyTextObjectV2) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherName, ok := other.(*NameOnlyTextObjectV2); ok {
		otherName.Name.CopyInto(n.Name)
	}
}

func (n *NameOnlyTextObjectV2) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameOnlyTextObjectV2) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameOnlyTextObjectV2) ToBytes(languageCode string) ([]byte, error) {
	var buf bytes.Buffer

	if err := models.WriteSegment(&buf, getSegment(n.Name.GetLocalizedContent(languageCode))); err != nil {
		return nil, fmt.Errorf("writing NameOnlyTextObjectV2 name segment: %w", err)
	}

	if len(n.unknown) > 0 {
		buf.Write(n.unknown)
	}
	return buf.Bytes(), nil
}

func (n *NameOnlyTextObjectV2) ToString(languageCode string) string {
	return n.GetName(languageCode)
}

func (n *NameOnlyTextObjectV2) String() string {
	return n.ToString(common.DefaultLocalization)
}
