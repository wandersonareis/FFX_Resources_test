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

type NameDescriptionEffectAbilityTextObject struct {
	Bytes          []byte
	Name           datastore.IGlobalLocalizedKeyedStringObject
	Description    datastore.IGlobalLocalizedKeyedStringObject
	Abilities      []datastore.IGlobalLocalizedKeyedStringObject
	Effect         datastore.IGlobalLocalizedKeyedStringObject
	EffectPosition int64
	HeaderLength   int
}

func NewNameDescriptionEffectAbilityTextObject(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	abilitiesCount int,
	effectSegmentPosition int64,
	languageCode string,
) (*NameDescriptionEffectAbilityTextObject, error) {
	if common.GetGameVersionString() != "ffx2" {
		return nil, fmt.Errorf("NameDescriptionEffectAbilityTextObject is only compatible with FFX-2")
	}

	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create NameDescriptionEffectAbilityTextObject: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &NameDescriptionEffectAbilityTextObject{
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

func (p *NameDescriptionEffectAbilityTextObject) mapBytes(stringBytes []byte, languageCode string) error {
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

func (p *NameDescriptionEffectAbilityTextObject) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(p.Bytes)

	sequential := make([]datastore.IGlobalLocalizedKeyedStringObject, 0, 2+len(p.Abilities))
	sequential = append(sequential,
		p.Name,
		p.Description,
	)
	sequential = append(sequential, p.Abilities...)

	if err := writeStringSegments(data, nameSegmentDefaultPosition, languageCode, sequential...); err != nil {
		return nil, err
	}

	if effectContent := p.Effect.GetLocalizedContent(languageCode); effectContent != nil {
		if err := models.WriteSegmentAt(data, int(p.EffectPosition), getSegment(effectContent)); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (p *NameDescriptionEffectAbilityTextObject) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *NameDescriptionEffectAbilityTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (p *NameDescriptionEffectAbilityTextObject) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *NameDescriptionEffectAbilityTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *NameDescriptionEffectAbilityTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*NameDescriptionEffectAbilityTextObject); ok {
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

func (p *NameDescriptionEffectAbilityTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
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

func (p *NameDescriptionEffectAbilityTextObject) ToString(languageCode string) string {
	nameStr := p.GetName(languageCode)
	descStr := ""
	if descContent := p.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	return fmt.Sprintf("%s - %s", nameStr, descStr)
}

func (p *NameDescriptionEffectAbilityTextObject) String() string {
	return p.ToString(common.DefaultLocalization)
}
