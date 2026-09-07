package objectsfile

import (
	"bytes"
	"fmt"
	"io"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type NameDescriptionEffect struct {
	Bytes                 []byte
	Name                  datastore.IGlobalLocalizedKeyedStringObject
	Description           datastore.IGlobalLocalizedKeyedStringObject
	Effect                datastore.IGlobalLocalizedKeyedStringObject
	EffectSegmentPosition int64
	HeaderLength          int
}

func NewNameDescriptionEffect(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	effectSegmentPosition int64,
	languageCode string,
) (*NameDescriptionEffect, error) {
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create NameDescriptionEffect: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &NameDescriptionEffect{
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

func (p *NameDescriptionEffect) mapBytes(
	stringBytes []byte,
	languageCode string,
) error {
	r := bytes.NewReader(p.Bytes)

	nameSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogError("Error reading NameDescriptionEffect name: %v", err)
		return err
	}
	p.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, nameSeg.Offset, nameSeg.Key)

	descSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogError("Error reading NameDescriptionEffect description: %v", err)
		return err
	}
	p.Description.ReadAndSetLocalizedContent(languageCode, stringBytes, descSeg.Offset, descSeg.Key)

	if _, err := r.Seek(p.EffectSegmentPosition, io.SeekStart); err != nil {
		common.LogError("Error seeking to NameDescriptionEffect effect: %v", err)
		return err
	}

	effectSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogError("Error reading NameDescriptionEffect effect: %v", err)
		return err
	}
	p.Effect.ReadAndSetLocalizedContent(languageCode, stringBytes, effectSeg.Offset, effectSeg.Key)

	return nil
}

func (p *NameDescriptionEffect) ToBytes(languageCode string) ([]byte, error) {
	data := make([]byte, len(p.Bytes))
	copy(data, p.Bytes)

	if nameContent := p.Name.GetLocalizedContent(languageCode); nameContent != nil {
		if err := models.WriteSegmentAt(data, nameSegmentDefaultPosition, getSegment(nameContent)); err != nil {
			return nil, err
		}
	}

	if descContent := p.Description.GetLocalizedContent(languageCode); descContent != nil {
		if err := models.WriteSegmentAt(data, descriptionSegmentDefaultPosition, getSegment(descContent)); err != nil {
			return nil, err
		}
	}

	if effectContent := p.Effect.GetLocalizedContent(languageCode); effectContent != nil {
		if err := models.WriteSegmentAt(data, int(p.EffectSegmentPosition), getSegment(effectContent)); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (p *NameDescriptionEffect) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *NameDescriptionEffect) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (p *NameDescriptionEffect) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *NameDescriptionEffect) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *NameDescriptionEffect) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameDescriptionEffect); ok {
		o.Name.CopyInto(p.Name)
		o.Description.CopyInto(p.Description)
		o.Effect.CopyInto(p.Effect)
	}
}

func (p *NameDescriptionEffect) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.Description.GetLocalizedContent(localization),
		p.Effect.GetLocalizedContent(localization),
	}
}

func (p *NameDescriptionEffect) ToString(languageCode string) string {
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

func (p *NameDescriptionEffect) String() string {
	return p.ToString(common.DefaultLocalization)
}
