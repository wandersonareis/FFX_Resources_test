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

type NameDescriptionEffectTextObject struct {
	Bytes                 []byte
	Name                  datastore.IGlobalLocalizedKeyedStringObject
	Description           datastore.IGlobalLocalizedKeyedStringObject
	Effect                datastore.IGlobalLocalizedKeyedStringObject
	EffectSegmentPosition int64
	HeaderLength          int
}

func NewNameDescriptionEffectTextObject(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	effectSegmentPosition int64,
	languageCode string,
) (*NameDescriptionEffectTextObject, error) {
	if common.GetGameVersionString() != "ffx2" {
		return nil, fmt.Errorf("NameDescriptionEffectTextObject is only compatible with FFX-2")
	}

	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create NameDescriptionEffectTextObject: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &NameDescriptionEffectTextObject{
		Bytes:                 bytes,
		Name:                  NewLocalizedKeyedStringObject(),
		Description:           NewLocalizedKeyedStringObject(),
		Effect:                NewLocalizedKeyedStringObject(),
		EffectSegmentPosition: effectSegmentPosition,
		HeaderLength:          headerLength,
	}

	if err := p.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *NameDescriptionEffectTextObject) mapBytes(
	stringBytes []byte,
	languageCode string,
) error {
	r := bytes.NewReader(p.Bytes)

	if err := readStringSegments(r, stringBytes, languageCode, p.Name, p.Description); err != nil {
		common.LogError("Error reading NameDescriptionEffectTextObject sequential segments: %v", err)
		return err
	}

	if _, err := r.Seek(p.EffectSegmentPosition, io.SeekStart); err != nil {
		common.LogError("Error seeking to NameDescriptionEffectTextObject effect: %v", err)
		return err
	}

	effectSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogError("Error reading NameDescriptionEffectTextObject effect: %v", err)
		return err
	}
	p.Effect.ReadAndSetLocalizedContent(languageCode, stringBytes, effectSeg.Offset, effectSeg.Key)

	return nil
}

func (p *NameDescriptionEffectTextObject) ToBytes(languageCode string) ([]byte, error) {
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

func (p *NameDescriptionEffectTextObject) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *NameDescriptionEffectTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (p *NameDescriptionEffectTextObject) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *NameDescriptionEffectTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *NameDescriptionEffectTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameDescriptionEffectTextObject); ok {
		o.Name.CopyInto(p.Name)
		o.Description.CopyInto(p.Description)
		o.Effect.CopyInto(p.Effect)
	}
}

func (p *NameDescriptionEffectTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.Description.GetLocalizedContent(localization),
		p.Effect.GetLocalizedContent(localization),
	}
}

func (p *NameDescriptionEffectTextObject) ToString(languageCode string) string {
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

func (p *NameDescriptionEffectTextObject) String() string {
	return p.ToString(common.DefaultLocalization)
}
