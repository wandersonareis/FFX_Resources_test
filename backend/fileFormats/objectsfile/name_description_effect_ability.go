package objectsfile

import (
	"bytes"
	"fmt"
	"io"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
)

type NameDescriptionEffectAbilityObjectV2 struct {
	Bytes          []byte
	Name           datastore.IGlobalLocalizedKeyedStringObject
	Description    datastore.IGlobalLocalizedKeyedStringObject
	Abilities      []datastore.IGlobalLocalizedKeyedStringObject
	Effect         datastore.IGlobalLocalizedKeyedStringObject
	EffectPosition int64
	HeaderLength   int
}

func NewNameDescriptionEffectAbility(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	abilitiesCount int,
	effectSegmentPosition int64,
	languageCode string,
) (*NameDescriptionEffectAbilityObjectV2, error) {
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create NameDescriptionEffectAbility: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &NameDescriptionEffectAbilityObjectV2{
		Bytes:          bytes,
		Name:           NewLocalizedKeyedStringObject(),
		Description:    NewLocalizedKeyedStringObject(),
		Effect:         NewLocalizedKeyedStringObject(),
		EffectPosition: effectSegmentPosition,
		Abilities:      make([]datastore.IGlobalLocalizedKeyedStringObject, abilitiesCount),
		HeaderLength:   headerLength,
	}

	for i := range p.Abilities {
		p.Abilities[i] = NewLocalizedKeyedStringObject()
	}

	if err := p.mapBytes(stringBytes, languageCode); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *NameDescriptionEffectAbilityObjectV2) mapBytes(stringBytes []byte, languageCode string) error {
	r := bytes.NewReader(p.Bytes)

	sequentialSegments := make([]datastore.IGlobalLocalizedKeyedStringObject, 0, 2+len(p.Abilities))
	sequentialSegments = append(sequentialSegments, p.Name, p.Description)
	sequentialSegments = append(sequentialSegments, p.Abilities...)

	if err := readStringSegments(r, stringBytes, languageCode, sequentialSegments...); err != nil {
		common.LogError("Error reading NameDescriptionEffectAbility sequential segments: %v", err)
		return err
	}

	if _, err := r.Seek(p.EffectPosition, io.SeekStart); err != nil {
		common.LogError("Error seeking to NameDescriptionEffectAbility effect: %v", err)
		return err
	}

	effectSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogError("Error reading NameDescriptionEffectAbility effect: %v", err)
		return err
	}
	p.Effect.ReadAndSetLocalizedContent(languageCode, stringBytes, effectSeg.Offset, effectSeg.Key)

	return nil
}

func (p *NameDescriptionEffectAbilityObjectV2) ToBytes(languageCode string) ([]byte, error) {
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

	// Abilities: sequenciais, começando imediatamente após a Description
	abilityStartPos := descriptionSegmentDefaultPosition + 4 // 4 bytes do segmento Description
	for i := range p.Abilities {
		if abilityContent := p.Abilities[i].GetLocalizedContent(languageCode); abilityContent != nil {
			pos := abilityStartPos + (i * 4) // Cada Ability tem 4 bytes
			if err := models.WriteSegmentAt(data, pos, getSegment(abilityContent)); err != nil {
				return nil, err
			}
		}
	}

	if effectContent := p.Effect.GetLocalizedContent(languageCode); effectContent != nil {
		if err := models.WriteSegmentAt(data, int(p.EffectPosition), getSegment(effectContent)); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (p *NameDescriptionEffectAbilityObjectV2) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *NameDescriptionEffectAbilityObjectV2) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return p.Name
	case "description":
		return p.Description
	case "effect":
		return p.Effect
	default:
		for i := range p.Abilities {
			if title == fmt.Sprintf("ability%d", i+1) {
				return p.Abilities[i]
			}
		}
		return nil
	}
}

func (p *NameDescriptionEffectAbilityObjectV2) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *NameDescriptionEffectAbilityObjectV2) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *NameDescriptionEffectAbilityObjectV2) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameDescriptionEffectAbilityObjectV2); ok {
		o.Name.CopyInto(p.Name)
		o.Description.CopyInto(p.Description)

		for i := range p.Abilities {
			if i < len(o.Abilities) {
				o.Abilities[i].CopyInto(p.Abilities[i])
			}
		}

		o.Effect.CopyInto(p.Effect)
	}
}

func (p *NameDescriptionEffectAbilityObjectV2) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	result := make([]datastore.IGlobalKeyedString, 0, 3+len(p.Abilities))
	result = append(result,
		p.Name.GetLocalizedContent(localization),
		p.Description.GetLocalizedContent(localization),
	)
	for _, seg := range p.Abilities {
		result = append(result, seg.GetLocalizedContent(localization))
	}
	result = append(result, p.Effect.GetLocalizedContent(localization))
	return result
}

func (p *NameDescriptionEffectAbilityObjectV2) ToString(languageCode string) string {
	nameStr := p.GetName(languageCode)
	descStr := ""
	if descContent := p.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	return fmt.Sprintf("%s - %s", nameStr, descStr)
}

func (p *NameDescriptionEffectAbilityObjectV2) String() string {
	return p.ToString(common.DefaultLocalization)
}
