package objectsfile

import (
	"bytes"
	"fmt"
	"slices"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

type LastMissionTextObject struct {
	Bytes                 []byte
	Name                  datastore.IGlobalLocalizedKeyedStringObject
	Description           datastore.IGlobalLocalizedKeyedStringObject
	Effect                datastore.IGlobalLocalizedKeyedStringObject
	EffectDescription     datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength          int
	Version               common.GameVersion
}

func NewLastMissionTextObject(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	languageCode string,
	version common.GameVersion,
) (*LastMissionTextObject, error) {
	if version.Normalize() != common.GameVersionLastMiss {
		return nil, fmt.Errorf("LastMissionTextObject is only compatible with LastMission, but got game version %s", version)
	}

	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create LastMissionTextObject: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &LastMissionTextObject{
		Bytes:             bytes,
		Name:              NewLocalizedKeyedStringObject(),
		Description:       NewLocalizedKeyedStringObject(),
		Effect:            NewLocalizedKeyedStringObject(),
		EffectDescription: NewLocalizedKeyedStringObject(),
		HeaderLength:      headerLength,
		Version:           version,
	}

	if err := p.mapBytes(stringBytes, languageCode, version); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *LastMissionTextObject) mapBytes(
	stringBytes []byte,
	languageCode string,
	version common.GameVersion,
) error {
	r := bytes.NewReader(p.Bytes)

	if err := readStringSegments(r, stringBytes, languageCode, version, p.Name, p.Description); err != nil {
		common.LogError("Error reading LastMissionTextObject sequential segments: %v", err)
		return err
	}

	if err := readStringSegments(r, stringBytes, languageCode, version, p.Effect, p.EffectDescription); err != nil {
		common.LogError("Error reading LastMissionTextObject effect segments: %v", err)
		return err
	}

	return nil
}

func (p *LastMissionTextObject) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(p.Bytes)

	if err := writeStringSegments(data, 0, languageCode, p.Name, p.Description, p.Effect, p.EffectDescription); err != nil {
		return nil, err
	}

	return data, nil
}

func (p *LastMissionTextObject) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *LastMissionTextObject) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return p.Name
	case "description":
		return p.Description
	case "effect":
		return p.Effect
	case "effectDescription":
		return p.EffectDescription
	default:
		return nil
	}
}

func (p *LastMissionTextObject) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *LastMissionTextObject) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *LastMissionTextObject) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*LastMissionTextObject); ok {
		o.Name.CopyInto(p.Name)
		o.Description.CopyInto(p.Description)
		o.Effect.CopyInto(p.Effect)
		o.EffectDescription.CopyInto(p.EffectDescription)
	}
}

func (p *LastMissionTextObject) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.Description.GetLocalizedContent(localization),
		p.Effect.GetLocalizedContent(localization),
		p.EffectDescription.GetLocalizedContent(localization),
	}
}

func (p *LastMissionTextObject) ToString(languageCode string) string {
	nameStr := p.GetName(languageCode)
	descStr := ""
	if descContent := p.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	effStr := ""
	if effContent := p.Effect.GetLocalizedContent(languageCode); effContent != nil {
		effStr = effContent.GetString()
	}
	effDescStr := ""
	if effDescContent := p.EffectDescription.GetLocalizedContent(languageCode); effDescContent != nil {
		effDescStr = effDescContent.GetString()
	}
	return fmt.Sprintf("%s - %s - %s - %s", nameStr, descStr, effStr, effDescStr)
}

func (p *LastMissionTextObject) String() string {
	return p.ToString(common.DefaultLocalization)
}