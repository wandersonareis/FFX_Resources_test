package objectsfile

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"io"
)

type (
	NameOnlyDataObject struct {
		Bytes            []byte
		Name             datastore.IGlobalLocalizedKeyedStringObject
		FirstSeparator   datastore.IGlobalLocalizedKeyedStringObject
		headerParameters []byte
		HeaderLength     int
	}
	/* nameOnlyHeaderData struct {
		NameOffset           uint16
		NameKey              uint16
		FirstSeparatorOffset uint16
		FirstSeparatorKey    uint16
	} */
)

type NameOnlyDataObjectV2 struct {
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

const NameOnlyDataObjectLength int = 0x10

func getValidHeader(data []byte, max int) []byte {
	const minRequired = 0x8

	if len(data) < minRequired {
		return nil
	}

	end := min(max, len(data))

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

func (n *NameOnlyDataObject) readHeaderData(reader *bytes.Reader) (models.NameOnlyHeaderData, error) {
	nameSegment, err := models.ReadSegment(reader)
	if err != nil {
		if err == io.EOF {
			panic(fmt.Errorf("unexpected EOF while reading NameOnlyDataObject header"))
		}
		return models.NameOnlyHeaderData{}, err
	}

	firstSeparator, err := models.ReadSegment(reader)
	if err != nil {
		if err == io.EOF {
			panic(fmt.Errorf("unexpected EOF while reading NameOnlyDataObject header"))
		}
		return models.NameOnlyHeaderData{}, err
	}

	return models.NameOnlyHeaderData{
		NameSegment:           nameSegment,
		FirstSeparatorSegment: firstSeparator,
	}, nil
}

func (n *NameOnlyDataObject) mapBytes(stringBytes []byte, languageCode string) {
	r := bytes.NewReader(getValidHeader(n.Bytes, NameOnlyDataObjectLength))

	hd, err := n.readHeaderData(r)
	if err != nil {
		fmt.Printf("Error reading NameOnlyDataObject: %v\n", err)
		return
	}
	n.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.NameSegment.Offset, hd.NameSegment.Key)
	n.FirstSeparator.ReadAndSetLocalizedContent(languageCode, stringBytes, hd.FirstSeparatorSegment.Offset, hd.FirstSeparatorSegment.Key)
}

func (n *NameOnlyDataObject) ToBytes(languageCode string) []byte {
	result := make([]byte, n.HeaderLength)

	var buf bytes.Buffer
/* 	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.Name.GetLocalizedContent(languageCode)))
	binary.Write(&buf, binary.LittleEndian, getLocalizedBytes(n.FirstSeparator.GetLocalizedContent(languageCode))) */
	models.WriteSegment(&buf, getSegment(n.Name.GetLocalizedContent(languageCode)))
	models.WriteSegment(&buf, getSegment(n.FirstSeparator.GetLocalizedContent(languageCode)))

	copy(result, buf.Bytes())

	if len(n.headerParameters) > 0 && NameOnlyDataObjectLength+len(n.headerParameters) <= n.HeaderLength {
		copy(result[NameOnlyDataObjectLength:], n.headerParameters)
	}

	return result
}

func (n *NameOnlyDataObject) ToList(filename string, languageCode string) components.IList[datastore.IGlobalLocalizedTextObject] {
	creator := func(data []byte, stringBytes []byte, headerLength int, loc string) datastore.IGlobalLocalizedTextObject {
		return NewNameOnlyDataObject(data, stringBytes, headerLength, loc)
	}
	return ReadDataListWithIlist(filename, languageCode, creator)
}

func (n *NameOnlyDataObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameOnlyDataObject) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *NameOnlyDataObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (n *NameOnlyDataObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherName, ok := other.(*NameOnlyDataObject); ok {
		otherName.Name.CopyInto(n.Name)
		otherName.FirstSeparator.CopyInto(n.FirstSeparator)
	}
}

func (n *NameOnlyDataObject) GetLocalizedKeyedStrings(languageCode string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
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

func NewNameOnlyDataObjectV2(data []byte, stringBytes []byte, headerLength int, languageCode string) *NameOnlyDataObjectV2 {
	if len(data) < 4 {
		common.LogVerbose("Insufficient data to create NameOnlyDataObjectV2!")
		return nil
	}
	n := &NameOnlyDataObjectV2{
		Bytes:        data,
		Name:         NewLocalizedKeyedStringObject(),
		HeaderLength: headerLength,
	}
	n.mapBytes(stringBytes, languageCode)
	return n
}

func (n *NameOnlyDataObjectV2) mapBytes(stringBytes []byte, languageCode string) {
	r := bytes.NewReader(n.Bytes)
	seg, err := models.ReadSegment(r)
	if err != nil {
		fmt.Printf("Error reading NameOnlyDataObjectV2 name: %v\n", err)
		return
	}
	n.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, seg.Offset, seg.Key)
	if r.Len() > 0 {
		n.unknown = make([]byte, r.Len())
		r.Read(n.unknown)
	}
}

func (n *NameOnlyDataObjectV2) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *NameOnlyDataObjectV2) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	default:
		return nil
	}
}

func (n *NameOnlyDataObjectV2) GetLocalizedKeyedStrings(languageCode string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Name.GetLocalizedContent(languageCode),
	}
}

func (n *NameOnlyDataObjectV2) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherName, ok := other.(*NameOnlyDataObjectV2); ok {
		otherName.Name.CopyInto(n.Name)
	}
}

func (n *NameOnlyDataObjectV2) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *NameOnlyDataObjectV2) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *NameOnlyDataObjectV2) ToBytes(languageCode string) []byte {
	var buf bytes.Buffer
	models.WriteSegment(&buf, getSegment(n.Name.GetLocalizedContent(languageCode)))
	if len(n.unknown) > 0 {
		buf.Write(n.unknown)
	}
	return buf.Bytes()
}

func (n *NameOnlyDataObjectV2) ToString(languageCode string) string {
	return n.GetName(languageCode)
}

func (n *NameOnlyDataObjectV2) String() string {
	return n.ToString(common.DefaultLocalization)
}
