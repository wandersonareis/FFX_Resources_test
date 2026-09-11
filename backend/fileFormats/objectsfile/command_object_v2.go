package objectsfile

import (
	"bytes"
	"fmt"
	"slices"

	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

type CommandTextObjectV2 struct {
	Bytes        []byte
	Name         datastore.IGlobalLocalizedKeyedStringObject
	Description  datastore.IGlobalLocalizedKeyedStringObject
	HeaderLength int
}

func NewCommandTextObjectV2(bytes []byte, stringBytes []byte, headerLength int, languageCode string, gameVersion int) (*CommandTextObjectV2, error) {
	if gameVersion != 2 {
		return nil, fmt.Errorf("CommandTextObjectV2 is only compatible with FFX-2 (game version 2), but got game version %d", gameVersion)
	}
	
	if len(bytes) < headerLength {
		return nil, fmt.Errorf("insufficient data: have %d bytes, need at least %d", len(bytes), headerLength)
	}

	n := &CommandTextObjectV2{
		Bytes:        bytes,
		Name:         NewLocalizedKeyedStringObject(),
		Description:  NewLocalizedKeyedStringObject(),
		HeaderLength: headerLength,
	}

	if err := n.mapBytes(stringBytes, languageCode, gameVersion); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *CommandTextObjectV2) mapBytes(stringBytes []byte, languageCode string, version int) error {
	r := bytes.NewReader(n.Bytes)

	if err := readStringSegments(r, stringBytes, languageCode, version, n.Name, n.Description); err != nil {
		common.LogError("Error reading CommandTextObjectV2 segments: %v", err)
		return err
	}
	return nil
}

func (n *CommandTextObjectV2) ToBytes(languageCode string) ([]byte, error) {
	result := slices.Clone(n.Bytes)

	if err := writeStringSegments(result, 0, languageCode,
		n.Name,
		n.Description,
	); err != nil {
		return nil, err
	}
	return result, nil
}

func (n *CommandTextObjectV2) GetName(languageCode string) string {
	return n.Name.GetLocalizedString(languageCode)
}

func (n *CommandTextObjectV2) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	switch title {
	case "name":
		return n.Name
	case "description":
		return n.Description
	default:
		return nil
	}
}

func (n *CommandTextObjectV2) GetHeaderLength() int {
	return n.HeaderLength
}

func (n *CommandTextObjectV2) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return n
}

func (n *CommandTextObjectV2) SetLocalizations(other datastore.IGlobalLocalizationSetter) {
	if otherCmd, ok := other.(*CommandTextObjectV2); ok {
		otherCmd.Name.CopyInto(n.Name)
		otherCmd.Description.CopyInto(n.Description)
	}
}

func (n *CommandTextObjectV2) GetLocalizedKeyedStrings(localization string) []datastore.IGlobalKeyedString {
	return []datastore.IGlobalKeyedString{
		n.Name.GetLocalizedContent(localization),
		n.Description.GetLocalizedContent(localization),
	}
}

func (n *CommandTextObjectV2) ToString(languageCode string) string {
	nameStr := n.GetName(languageCode)
	descStr := ""
	if descContent := n.Description.GetLocalizedContent(languageCode); descContent != nil {
		descStr = descContent.GetString()
	}
	return fmt.Sprintf("%s - %s", nameStr, descStr)
}

func (n *CommandTextObjectV2) String() string {
	return n.ToString(common.DefaultLocalization)
}