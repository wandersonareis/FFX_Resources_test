package event

import (
	"ffxresources/backend/common"
	"strings"
)

type EventFile struct {
	ID      string
	Version common.GameVersion
	Strings []*LocalizedFieldStringObject
}

func NewEventFile(id string, version common.GameVersion, strings []*LocalizedFieldStringObject) *EventFile {
	return &EventFile{
		ID:      id,
		Version: version,
		Strings: strings,
	}
}

func (ef *EventFile) AddLocalizations(strings []*LocalizedFieldStringObject) {
	if ef.Strings == nil {
		ef.Strings = strings
		return
	}

	for i, localizationStringObject := range strings {
		if i < len(ef.Strings) {
			stringObject := ef.Strings[i]
			if stringObject != nil && localizationStringObject != nil {
				localizationStringObject.CopyInto(stringObject)
			}
		} else {
			ef.Strings = append(ef.Strings, localizationStringObject)
		}
	}
}

func (ef *EventFile) String() string {
	var builder strings.Builder
	builder.WriteString(ef.ID)
	builder.WriteString("\n")
	return builder.String()
}

func (ef *EventFile) GetID() string {
	return ef.ID
}

