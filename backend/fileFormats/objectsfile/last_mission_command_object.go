package objectsfile

import (
	"fmt"
	"slices"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

type LastMissionCommand struct {
	Bytes       []byte
	Name        datastore.IGlobalLocalizedKeyedStringObject
	Description datastore.IGlobalLocalizedKeyedStringObject
	Effect      datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength int
	Version      common.GameVersion
	skip        int
}

func NewLastMissionCommand(
	bytes []byte,
	stringBytes []byte,
	headerLength int,
	skip int,
	languageCode string,
	version common.GameVersion,
) (*LastMissionCommand, error) {
	if version.Normalize() != common.GameVersionLastMiss {
		return nil, fmt.Errorf("LastMissionCommand is only compatible with LastMission, but got game version %s", version)
	}

	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data to create LastMissionCommand: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	p := &LastMissionCommand{
		Bytes:        bytes,
		Name:         NewLocalizedKeyedStringObject(),
		Description:  NewLocalizedKeyedStringObject(),
		Effect:       NewLocalizedKeyedStringObject(),
		HeaderLength: headerLength,
		Version:      version,
		skip:         skip,
	}

	if err := p.mapBytes(stringBytes, languageCode, version); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *LastMissionCommand) offsets() []int {
	return segmentOffsets(0, []int{p.skip, p.skip}, 3)
}

func (p *LastMissionCommand) mapBytes(stringBytes []byte, languageCode string, version common.GameVersion) error {
	if err := readStringSegmentsAt(p.Bytes, p.offsets(), stringBytes, languageCode, version, p.Name, p.Description, p.Effect); err != nil {
		common.LogError("Error reading LastMissionCommand segments: %v", err)
		return err
	}
	return nil
}

func (p *LastMissionCommand) ToBytes(languageCode string) ([]byte, error) {
	data := slices.Clone(p.Bytes)
	if err := writeStringSegmentsAt(data, p.offsets(), languageCode, p.Name, p.Description, p.Effect); err != nil {
		return nil, err
	}
	return data, nil
}

func (p *LastMissionCommand) GetName(languageCode string) string {
	return p.Name.GetLocalizedString(languageCode)
}

func (p *LastMissionCommand) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
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

func (p *LastMissionCommand) GetHeaderLength() int {
	return p.HeaderLength
}

func (p *LastMissionCommand) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return p
}

func (p *LastMissionCommand) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if o, ok := other.(*LastMissionCommand); ok {
		o.Name.CopyInto(p.Name)
		o.Description.CopyInto(p.Description)
		o.Effect.CopyInto(p.Effect)
	}
}

func (p *LastMissionCommand) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		p.Name.GetLocalizedContent(localization),
		p.Description.GetLocalizedContent(localization),
		p.Effect.GetLocalizedContent(localization),
	}
}

func (p *LastMissionCommand) ToString(languageCode string) string {
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

func (p *LastMissionCommand) String() string {
	return p.ToString(common.DefaultLocalization)
}