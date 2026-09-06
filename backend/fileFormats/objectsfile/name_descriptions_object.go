package objectsfile

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
)

/* type (
	ILocalizedTextObject interface {
		GetName(languageCode string) string
		GetKeyedString(title string) *localization.LocalizedKeyedStringObject
		GetLocalizedKeyedStrings(languageCode string) []*models.KeyedString
		SetLocalizations(other components.LocalizationSetter)
		GetTextObject() ILocalizedTextObject
		GetHeaderLength() int
		ToBytes(languageCode string) []byte
		ToString(languageCode string) string
		String() string
	}
) */

var (
	//KEY_ITEMS components.IList[datastore.IGlobalLocalizedTextObject]
	//COMMANDS  components.IList[datastore.IGlobalLocalizedTextObject]
	//ITEMS     components.IList[datastore.IGlobalLocalizedTextObject]
	MONMAGIC1 components.IList[datastore.IGlobalLocalizedTextObject]
	MONMAGIC2 components.IList[datastore.IGlobalLocalizedTextObject]

	// FFX-2 (v2) kernel objects (name+description format).
	A_ABILITY   components.IList[datastore.IGlobalLocalizedTextObject]
	ACCESSORY   components.IList[datastore.IGlobalLocalizedTextObject]
	BATTLE_TEXT   components.IList[datastore.IGlobalLocalizedTextObject]
	JOB         components.IList[datastore.IGlobalLocalizedTextObject]
	MONMAGIC    components.IList[datastore.IGlobalLocalizedTextObject]
	MONSTER     components.IList[datastore.IGlobalLocalizedTextObject]
	MONSTER2    components.IList[datastore.IGlobalLocalizedTextObject]
	OVERSOUL    components.IList[datastore.IGlobalLocalizedTextObject]
	PLATE       components.IList[datastore.IGlobalLocalizedTextObject]
	PLAYER_SAVE components.IList[datastore.IGlobalLocalizedTextObject]
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
		Bytes            []byte
		Name             datastore.IGlobalLocalizedKeyedStringObject
		FirstSeparator   datastore.IGlobalLocalizedKeyedStringObject
		Description      datastore.IGlobalLocalizedKeyedStringObject
		SecondSeparator  datastore.IGlobalLocalizedKeyedStringObject
		headerParameters []byte
		HeaderLength     int
	}
)

const NameDescriptionTextObjectLength = 0x10

func NewNameDescriptionTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string) *NameDescriptionTextObject {
	if len(bytes) < headerLength {
		common.LogVerbose("Insufficient data to create NameDescriptionTextObject!")
		return nil
	}

	n := &NameDescriptionTextObject{
		Bytes:           bytes,
		Name:            NewLocalizedKeyedStringObject(),
		FirstSeparator:  NewLocalizedKeyedStringObject(),
		Description:     NewLocalizedKeyedStringObject(),
		SecondSeparator: NewLocalizedKeyedStringObject(),
		HeaderLength:    headerLength,
	}
	n.mapBytes(stringBytes, languageCode)
	return n
}

func (n *NameDescriptionTextObject) readHeaderData(reader *bytes.Reader) (models.NameDescriptionHeaderData, error) {
    nameSegment, err := models.ReadSegment(reader)
    if err != nil {
        return models.NameDescriptionHeaderData{}, err
    }

    firstSeparator, err := models.ReadSegment(reader)
    if err != nil {
        return models.NameDescriptionHeaderData{}, err
    }

    descriptionSegment, err := models.ReadSegment(reader)
    if err != nil {
        return models.NameDescriptionHeaderData{}, err
    }

    secondSeparator, err := models.ReadSegment(reader)
    if err != nil {
        return models.NameDescriptionHeaderData{}, err
    }

    return models.NameDescriptionHeaderData{
        NameSegment:            nameSegment,
        FirstSeparatorSegment:  firstSeparator,
        DescriptionSegment:     descriptionSegment,
        SecondSeparatorSegment: secondSeparator,
    }, nil
}

func (n *NameDescriptionTextObject) mapBytes(stringBytes []byte, languageCode string) {
	r := bytes.NewReader(getValidHeader(n.Bytes, NameDescriptionTextObjectLength))

	hd, err := n.readHeaderData(r)
	if err != nil {
		fmt.Printf("Error reading NameDescriptionTextObject: %v\n", err)
		return
	}

	n.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.NameSegment.Offset, hd.NameSegment.Key)
	n.FirstSeparator.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.FirstSeparatorSegment.Offset, hd.FirstSeparatorSegment.Key)
	n.Description.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.DescriptionSegment.Offset, hd.DescriptionSegment.Key)
	n.SecondSeparator.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.SecondSeparatorSegment.Offset, hd.SecondSeparatorSegment.Key)

	if n.HeaderLength > NameDescriptionTextObjectLength {
		n.headerParameters = n.Bytes[NameDescriptionTextObjectLength:n.HeaderLength]
	}
}

func (n *NameDescriptionTextObject) ToBytes(languageCode string) []byte {
	var buf bytes.Buffer
 	models.WriteSegment(&buf, getSegment(n.Name.GetLocalizedContent(languageCode)))
	models.WriteSegment(&buf, getSegment(n.FirstSeparator.GetLocalizedContent(languageCode)))
	models.WriteSegment(&buf, getSegment(n.Description.GetLocalizedContent(languageCode)))
	models.WriteSegment(&buf, getSegment(n.SecondSeparator.GetLocalizedContent(languageCode)))

	if len(n.headerParameters) > 0 && NameDescriptionTextObjectLength+len(n.headerParameters) <= n.HeaderLength {
		buf.Write(n.headerParameters)
	}
	return buf.Bytes()
}

func (n *NameDescriptionTextObject) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (d *NameDescriptionTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (n *NameDescriptionTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameDescriptionTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherNameDesc, ok := other.(*NameDescriptionTextObject); ok {
		otherNameDesc.Name.CopyInto(n.Name)
		otherNameDesc.FirstSeparator.CopyInto(n.FirstSeparator)
		otherNameDesc.Description.CopyInto(n.Description)
		otherNameDesc.SecondSeparator.CopyInto(n.SecondSeparator)
	}
}

func (n *NameDescriptionTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
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


