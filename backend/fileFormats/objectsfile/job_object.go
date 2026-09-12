package objectsfile

import (
	"bytes"
	"fmt"
	"io"
	"slices"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type JobTextObject struct {
	Bytes                 []byte
	Name                  datastore.IGlobalLocalizedKeyedStringObject
	Description           datastore.IGlobalLocalizedKeyedStringObject
	Effect                datastore.IGlobalLocalizedKeyedStringObject
	EffectSegmentPosition int64
	HeaderLength          int
	Version               common.GameVersion
}

func NewJobTextObject(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	effectSegmentPosition int64,
	languageCode string,
	version common.GameVersion,
) (*JobTextObject, error) {
	if version.Normalize() != common.GameVersionFFX2 && version.Normalize() != common.GameVersionLastMiss {
		return nil, fmt.Errorf("JobTextObject is only compatible with FFX-2 (ffx2), but got game version %s", version)
	}

	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create JobTextObject: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &JobTextObject{
		Bytes:                 bytes,
		Name:                  NewLocalizedKeyedStringObject(),
		Description:           NewLocalizedKeyedStringObject(),
		Effect:                NewLocalizedKeyedStringObject(),
		EffectSegmentPosition: effectSegmentPosition,
		HeaderLength:          headerLength,
		Version:               version,
	}

	if err := p.mapBytes(stringBytes, languageCode, version); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *JobTextObject) mapBytes(
	stringBytes []byte,
	languageCode string,
	version common.GameVersion,
) error {
	r := bytes.NewReader(p.Bytes)

	if err := readStringSegments(r, stringBytes, languageCode, version, p.Name, p.Description); err != nil {
		common.LogError("Error reading JobTextObject sequential segments: %v", err)
		return err
	}

	if _, err := r.Seek(p.EffectSegmentPosition, io.SeekStart); err != nil {
		common.LogError("Error seeking to JobTextObject effect: %v", err)
		return err
	}

	effectSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogError("Error reading JobTextObject effect: %v", err)
		return err
	}
	p.Effect.ReadAndSetLocalizedContent(languageCode, stringBytes, effectSeg.Offset, effectSeg.Key, version)

	return nil
}

func (p *JobTextObject) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(p.Bytes)

	if err := writeStringSegments(data, 0, languageCode, p.Name, p.Description); err != nil {
		return nil, err
	}

	if effectContent := p.Effect.GetLocalizedContent(languageCode); effectContent != nil {
		if err := models.WriteSegmentAt(data, int(p.EffectSegmentPosition), getSegment(effectContent)); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (p *JobTextObject) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *JobTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return p.Name
	case "description":
		return p.Description
	case "effect":
		return p.Effect
	default:
		return nil
	}
}

func (p *JobTextObject) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *JobTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *JobTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*JobTextObject); ok {
		o.Name.CopyInto(p.Name)
		o.Description.CopyInto(p.Description)
		o.Effect.CopyInto(p.Effect)
	}
}

func (p *JobTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.Description.GetLocalizedContent(localization),
		p.Effect.GetLocalizedContent(localization),
	}
}

func (p *JobTextObject) ToString(languageCode string) string {
	nameStr := p.GetName(languageCode)
	descStr := ""
	if descContent := p.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	effStr := ""
	if effContent := p.Effect.GetLocalizedContent(languageCode); effContent != nil {
		effStr = effContent.GetString()
	}
	return fmt.Sprintf("%s - %s - %s", nameStr, descStr, effStr)
}

func (p *JobTextObject) String() string {
	return p.ToString(common.DefaultLocalization)
}