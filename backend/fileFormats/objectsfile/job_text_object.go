package objectsfile

import (
	"bytes"
	"fmt"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type JobTextObject struct {
	Bytes              []byte
	NameSegment        datastore.IGlobalLocalizedKeyedStringObject
	DescriptionSegment datastore.IGlobalLocalizedKeyedStringObject
	EffectSegment      datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength       int
}

const JobTextObjectLength = 0x0C

func NewJobTextObject(bytes []byte, stringBytes []byte, headerLength int, languageCode string) *JobTextObject {
	if len(bytes) < JobTextObjectLength {
		common.LogVerbose("Insufficient data to create JobTextObject!")
		return nil
	}

	j := &JobTextObject{
		Bytes:              bytes,
		NameSegment:        NewLocalizedKeyedStringObject(),
		DescriptionSegment: NewLocalizedKeyedStringObject(),
		EffectSegment:      NewLocalizedKeyedStringObject(),
		HeaderLength:       headerLength,
	}
	j.mapBytes(stringBytes, languageCode)
	return j
}

func (j *JobTextObject) mapBytes(stringBytes []byte, languageCode string) {
	r := bytes.NewReader(j.Bytes)

	nameSeg, err := models.ReadSegment(r)
	if err != nil {
		fmt.Printf("Error reading JobTextObject name: %v\n", err)
		return
	}
	descSeg, err := models.ReadSegment(r)
	if err != nil {
		fmt.Printf("Error reading JobTextObject description: %v\n", err)
		return
	}
	effSeg, err := models.ReadSegment(r)
	if err != nil {
		fmt.Printf("Error reading JobTextObject effect: %v\n", err)
		return
	}

	j.NameSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, nameSeg.Offset, nameSeg.Key)
	j.DescriptionSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, descSeg.Offset, descSeg.Key)

	if effSeg.Offset == 0xFF {
		descData := stringBytes[descSeg.Offset:]
		nullIdx := bytes.IndexByte(descData, 0x00)
		if nullIdx < 0 {
			j.EffectSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, descSeg.Offset+1, effSeg.Key)
		} else {
			realOffset := descSeg.Offset + models.Offset(nullIdx) + 1
			j.EffectSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, realOffset, effSeg.Key)
		}
	} else {
		j.EffectSegment.ReadAndSetLocalizedContent(languageCode, stringBytes, effSeg.Offset, effSeg.Key)
	}
}

func (j *JobTextObject) ToBytes(languageCode string) []byte {
	var buf bytes.Buffer
	models.WriteSegment(&buf, getSegment(j.NameSegment.GetLocalizedContent(languageCode)))
	models.WriteSegment(&buf, getSegment(j.DescriptionSegment.GetLocalizedContent(languageCode)))
	effSeg := getSegment(j.EffectSegment.GetLocalizedContent(languageCode))
	effSeg.Offset = 0xFF
	models.WriteSegment(&buf, effSeg)
	return buf.Bytes()
}

func (j *JobTextObject) GetName(languageCode string) string {
	return j.NameSegment.GetLocalizedString(languageCode)
}

func (j *JobTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return j.NameSegment
	case "description":
		return j.DescriptionSegment
	case "effect":
		return j.EffectSegment
	default:
		return nil
	}
}

func (j *JobTextObject) GetHeaderLength() int {
	return j.HeaderLength
}

func (j *JobTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return j
}

func (j *JobTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*JobTextObject); ok {
		o.NameSegment.CopyInto(j.NameSegment)
		o.DescriptionSegment.CopyInto(j.DescriptionSegment)
		o.EffectSegment.CopyInto(j.EffectSegment)
	}
}

func (j *JobTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		j.NameSegment.GetLocalizedContent(localization),
		j.DescriptionSegment.GetLocalizedContent(localization),
		j.EffectSegment.GetLocalizedContent(localization),
	}
}

func (j *JobTextObject) ToString(languageCode string) string {
	nameStr := j.GetName(languageCode)
	descStr := ""
	if descContent := j.DescriptionSegment.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	effStr := ""
	if effContent := j.EffectSegment.GetLocalizedContent(languageCode); effContent != nil {
		effStr = effContent.GetString()
	}
	return fmt.Sprintf("%s - %s - %s", nameStr, descStr, effStr)
}

func (j *JobTextObject) String() string {
	return j.ToString(common.DefaultLocalization)
}