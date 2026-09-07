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

var (
	MONMAGIC1 components.IList[datastore.IGlobalLocalizedTextObject]
	MONMAGIC2 components.IList[datastore.IGlobalLocalizedTextObject]

	A_ABILITY    components.IList[datastore.IGlobalLocalizedTextObject]
	ACCESSORY    components.IList[datastore.IGlobalLocalizedTextObject]
	BATTLE_TEXT  components.IList[datastore.IGlobalLocalizedTextObject]
	JOB          components.IList[datastore.IGlobalLocalizedTextObject]
	MONMAGIC     components.IList[datastore.IGlobalLocalizedTextObject]
	MONSTER      components.IList[datastore.IGlobalLocalizedTextObject]
	MONSTER2     components.IList[datastore.IGlobalLocalizedTextObject]
	OVERSOUL     components.IList[datastore.IGlobalLocalizedTextObject]
	PLATE        components.IList[datastore.IGlobalLocalizedTextObject]
	PLAYER_SAVE  components.IList[datastore.IGlobalLocalizedTextObject]
)

func GetCommand(idx int) datastore.IGlobalLocalizedTextObject {
	if datastore.Commands.IsEmpty() || idx >= datastore.Commands.Len() {
		return nil
	}
	return datastore.Commands.Get(idx)
}

type Nameable interface {
	GetName(string) string
}

func GetKeyItem(idx int) datastore.IGlobalLocalizedTextObject {
	if datastoreItem := datastore.KeyItems.Get(idx); datastoreItem != nil {
		return datastoreItem
	}
	return nil
}

func GetNameableObject(typ string, idx int) Nameable {
	switch typ {
	case "command":
		if cmd := GetCommand(idx); cmd != nil {
			return cmd
		}
	/* case "monster":
	if m := GetMonster(idx); m != nil {
		return m
	} */
	case "keyItem":
		if k := GetKeyItem(idx); k != nil {
			return k
		}
		/* case "treasure":
		if t := GetTreasure(idx); t != nil {
			return t
		} */
		/* case "sgNodeType":
		if s := GetSgNodeType(idx); s != nil {
			return s
		}*/
	}
	return nil
}

func getSegment(keyedObj datastore.IGlobalKeyedString) models.Segment {
	if keyedObj == nil {
		return models.Segment{Offset: 0, Key: 0}
	}
	if ks, ok := keyedObj.(*KeyedString); ok {
		return ks.Segment
	}
	return models.Segment{Offset: 0, Key: 0}
}

type (
	NameDescriptionTextObject struct {
		Bytes                  []byte
		Name                   datastore.IGlobalLocalizedKeyedStringObject
		SimplifiedName         datastore.IGlobalLocalizedKeyedStringObject
		Description            datastore.IGlobalLocalizedKeyedStringObject
		SimplifiedDescription  datastore.IGlobalLocalizedKeyedStringObject
		HeaderLength           int
	}
)

func NewNameDescriptionTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string) (*NameDescriptionTextObject, error) {
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	n := &NameDescriptionTextObject{
		Bytes:                 bytes,
		Name:                  NewLocalizedKeyedStringObject(),
		SimplifiedName:        NewLocalizedKeyedStringObject(),
		Description:           NewLocalizedKeyedStringObject(),
		SimplifiedDescription: NewLocalizedKeyedStringObject(),
		HeaderLength:          headerLength,
	}

	if err := n.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *NameDescriptionTextObject) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(n.Bytes)
	return readStringSegments(r, stringBytes, languageCode,
		n.Name,
		n.SimplifiedName,
		n.Description,
		n.SimplifiedDescription,
	)
}

func (n *NameDescriptionTextObject) ToBytes(languageCode string) ([]byte, error) {
	result := slices.Clone(n.Bytes)

	if err := writeStringSegments(result, []datastore.IGlobalLocalizedKeyedStringObject{
		n.Name,
		n.SimplifiedName,
		n.Description,
		n.SimplifiedDescription,
	}, languageCode); err != nil {
		return nil, err
	}

	return result, nil
}

func (n *NameDescriptionTextObject) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (d *NameDescriptionTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return d.Name
	case "simplifiedName":
		return d.SimplifiedName
	case "description":
		return d.Description
	case "simplifiedDescription":
		return d.SimplifiedDescription
	default:
		return nil
	}
}

func (n *NameDescriptionTextObject) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameDescriptionTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameDescriptionTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherNameDesc, ok := other.(*NameDescriptionTextObject); ok {
		otherNameDesc.Name.CopyInto(n.Name)
		otherNameDesc.SimplifiedName.CopyInto(n.SimplifiedName)
		otherNameDesc.Description.CopyInto(n.Description)
		otherNameDesc.SimplifiedDescription.CopyInto(n.SimplifiedDescription)
	}
}

func (n *NameDescriptionTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Name.GetLocalizedContent(localization),
		n.SimplifiedName.GetLocalizedContent(localization),
		n.Description.GetLocalizedContent(localization),
		n.SimplifiedDescription.GetLocalizedContent(localization),
	}
}

func (d *NameDescriptionTextObject) ToString(languageCode string) string {
	nameStr := d.GetName(languageCode)
	simplifiedNameStr := ""
	if simplifiedNameContent := d.SimplifiedName.GetLocalizedContent(languageCode); simplifiedNameContent != nil {
		simplifiedNameStr = simplifiedNameContent.GetString()
	}
	descStr := ""
	if descContent := d.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	simplifiedDescStr := ""
	if simplifiedDescContent := d.SimplifiedDescription.GetLocalizedContent(languageCode); simplifiedDescContent != nil {
		simplifiedDescStr = simplifiedDescContent.GetString()
	}
	return fmt.Sprintf("%s %s - %s %s", nameStr, simplifiedNameStr, descStr, simplifiedDescStr)
}

func (n *NameDescriptionTextObject) String() string {
	return n.ToString(common.DefaultLocalization)
}


