package event

import (
	"bytes"
	"ffxresources/backend/common"
	ffxencoding "ffxresources/backend/core/encoding"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"strings"
)

// EventKeyedStringFile é o objeto events que implementa IGlobalLocalizedTextObject
// para o ciclo de vida via IBinaryFile. Event-específico: não reutiliza o layout
// de objectsfile.
type EventKeyedStringFile struct {
	Info    *models.EventFileInfo
	Strings []*LocalizedFieldStringObject
}

func NewEventKeyedStringFile(info *models.EventFileInfo, strings []*LocalizedFieldStringObject) *EventKeyedStringFile {
	return &EventKeyedStringFile{Info: info, Strings: strings}
}

func (e *EventKeyedStringFile) GetName(languageCode string) string {
	for _, s := range e.Strings {
		if s != nil {
			if v := s.GetLocalizedString(languageCode); v != "" {
				return v
			}
		}
	}
	return ""
}

func (e *EventKeyedStringFile) GetKeyedString(title string) datastore.IGlobalLocalizedKeyedStringObject {
	return nil
}

func (e *EventKeyedStringFile) GetLocalizedKeyedStrings(languageCode string) []datastore.IGlobalKeyedString {
	return nil
}

func (e *EventKeyedStringFile) SetLocalizations(other datastore.IGlobalLocalizationSetter) {}

func (e *EventKeyedStringFile) GetTextObject() datastore.IGlobalLocalizedTextObject {
	return e
}

func (e *EventKeyedStringFile) GetHeaderLength() int {
	return len(e.Strings) * 8
}

func (e *EventKeyedStringFile) ToBytes(languageCode string) ([]byte, error) {
	if len(e.Strings) == 0 {
		return []byte{}, nil
	}

	charset := ffxencoding.GetCharsetForLanguage(languageCode)
	fieldStrings := make([]*FieldString, 0, len(e.Strings))
	for _, localizedObj := range e.Strings {
		if localizedObj == nil {
			continue
		}
		fieldString := localizedObj.GetLocalizedContent(languageCode)
		if fieldString == nil {
			fieldString = NewEmptyFieldString(charset, e.Info.Version)
		}
		fieldStrings = append(fieldStrings, fieldString)
	}

	stringBytes := RebuildFieldStrings(fieldStrings, charset, e.Info.Version)

	var buf bytes.Buffer
	for _, str := range fieldStrings {
		if str != nil {
			buf.Write(str.ToRegularHeaderBytes())
			buf.Write(str.ToSimplifiedHeaderBytes())
		}
	}
	buf.Write(stringBytes)
	return buf.Bytes(), nil
}

func (e *EventKeyedStringFile) ToString(languageCode string) string {
	var parts []string
	for _, s := range e.Strings {
		if s != nil {
			if v := s.GetLocalizedString(languageCode); v != "" {
				parts = append(parts, v)
			}
		}
	}
	return strings.Join(parts, "\n")
}

func (e *EventKeyedStringFile) String() string {
	return e.ToString(common.DefaultLocalization)
}