package objectsfile

import (
	"bytes"
	"fmt"

	"io"
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type NameDescriptionTextObjectV2 struct {
	Bytes              []byte
	NameSegment        datastore.IGlobalLocalizedKeyedStringObject
	DescriptionSegment datastore.IGlobalLocalizedKeyedStringObject
	unknownBytes       []byte
	HeaderLength       int
}

const NameDescriptionTextObjectV2Length = 0x08

func NewNameDescriptionTextObjectV2(bytes []byte, stringBytes []byte, headerLength int, languageCode string) *NameDescriptionTextObjectV2 {
	if len(bytes) < NameDescriptionTextObjectV2Length {
		common.LogVerbose("Insufficient data to create NameDescriptionTextObjectV2!")
		return nil
	}

	n := &NameDescriptionTextObjectV2{
		Bytes:              bytes,
		NameSegment:        NewLocalizedKeyedStringObject(),
		DescriptionSegment: NewLocalizedKeyedStringObject(),
		HeaderLength:       headerLength,
	}
	n.mapBytesV2(stringBytes, languageCode)
	return n
}

func (n *NameDescriptionTextObjectV2) mapBytesV2(stringBytes []byte, languageCode string) {
	r := bytes.NewReader(n.Bytes)

	nameSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameDescriptionTextObjectV2 name: %v", err)
		return
	}
	descSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameDescriptionTextObjectV2 description: %v", err)
		return
	}

	n.NameSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, nameSeg.Offset, nameSeg.Key)
	n.DescriptionSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, descSeg.Offset, descSeg.Key)

	if r.Len() > 0 {
		n.unknownBytes = make([]byte, r.Len())
		if _, err := io.ReadFull(r, n.unknownBytes); err != nil {
			common.LogVerbose("Error reading V2 unknown bytes: %v", err)
			n.unknownBytes = nil
		}
	}
}

func (n *NameDescriptionTextObjectV2) ToBytes(languageCode string) []byte {
	var buf bytes.Buffer
	models.WriteSegment(&buf, getSegment(n.NameSegment.GetLocalizedContent(languageCode)))
	models.WriteSegment(&buf, getSegment(n.DescriptionSegment.GetLocalizedContent(languageCode)))

	if len(n.unknownBytes) > 0 {
		buf.Write(n.unknownBytes)
	}
	return buf.Bytes()
}

func (n *NameDescriptionTextObjectV2) GetName(languageCode string) string {
	return n.NameSegment.GetLocalizedString(languageCode)
}

func (n *NameDescriptionTextObjectV2) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.NameSegment
	case "description":
		return n.DescriptionSegment
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
		o.NameSegment.CopyInto(n.NameSegment)
		o.DescriptionSegment.CopyInto(n.DescriptionSegment)
		if len(o.unknownBytes) > 0 {
			n.unknownBytes = append([]byte{}, o.unknownBytes...)
		}
	}
}

func (n *NameDescriptionTextObjectV2) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.NameSegment.GetLocalizedContent(localization),
		n.DescriptionSegment.GetLocalizedContent(localization),
	}
}

func (n *NameDescriptionTextObjectV2) ToString(languageCode string) string {
	nameStr := n.GetName(languageCode)
	descStr := ""
	if descContent := n.DescriptionSegment.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	return fmt.Sprintf("%s - %s", nameStr, descStr)
}

func (n *NameDescriptionTextObjectV2) String() string {
	return n.ToString(common.DefaultLocalization)
}