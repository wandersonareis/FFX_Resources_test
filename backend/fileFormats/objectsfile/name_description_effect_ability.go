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
) *NameDescriptionEffectAbilityObjectV2 {
	if len(bytes) < headerLength {
		common.LogVerbose("Insufficient data to create NameDescriptionEffectAbility!")
		return nil
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

	p.mapBytes(stringBytes, languageCode)

	return p
}

func (p *NameDescriptionEffectAbilityObjectV2) mapBytes(stringBytes []byte, languageCode string) {
	r := bytes.NewReader(p.Bytes)

	nameSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameDescriptionEffectAbility name: %v", err)
		return
	}
	p.Name.ReadAndSetLocalizedContent(languageCode, stringBytes, nameSeg.Offset, nameSeg.Key)

	descSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameDescriptionEffectAbility description: %v", err)
		return
	}
	p.Description.ReadAndSetLocalizedContent(languageCode, stringBytes, descSeg.Offset, descSeg.Key)

	for i := range p.Abilities {
		abSeg, err := models.ReadSegment(r)
		if err != nil {
			common.LogVerbose("Error reading NameDescriptionEffectAbility ability %d: %v", i+1, err)
			continue
		}
		p.Abilities[i].ReadAndSetLocalizedContent(languageCode, stringBytes, abSeg.Offset, abSeg.Key)
	}

	if _, err := r.Seek(p.EffectPosition, io.SeekStart); err != nil {
		common.LogVerbose("Error seeking to NameDescriptionEffectAbility effect: %v", err)
		return
	}

	effectSeg, err := models.ReadSegment(r)
	if err != nil {
		common.LogVerbose("Error reading NameDescriptionEffectAbility effect: %v", err)
		return
	}
	p.Effect.ReadAndSetLocalizedContent(languageCode, stringBytes, effectSeg.Offset, effectSeg.Key)
}

func (p *NameDescriptionEffectAbilityObjectV2) ToBytes(languageCode string) []byte {
	data := make([]byte, len(p.Bytes))
	copy(data, p.Bytes)

	if nameContent := p.Name.GetLocalizedContent(languageCode); nameContent != nil {
		models.WriteSegmentAt(data, nameSegmentDefaultPosition, getSegment(nameContent))
	}

	if descContent := p.Description.GetLocalizedContent(languageCode); descContent != nil {
		models.WriteSegmentAt(data, descriptionSegmentDefaultPosition, getSegment(descContent))
	}

	// Abilities: sequenciais, começando imediatamente após a Description
	abilityStartPos := descriptionSegmentDefaultPosition + 4 // 4 bytes do segmento Description
	for i := range p.Abilities {
		if abilityContent := p.Abilities[i].GetLocalizedContent(languageCode); abilityContent != nil {
			pos := abilityStartPos + (i * 4) // Cada Ability tem 4 bytes
			models.WriteSegmentAt(data, pos, getSegment(abilityContent))
		}
	}

	if effectContent := p.Effect.GetLocalizedContent(languageCode); effectContent != nil {
		models.WriteSegmentAt(data, int(p.EffectPosition), getSegment(effectContent))
	}

	return data
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
