package components

import (
	"ffxresources/backend/common"
	"fmt"
	"strings"
)

type LocalizedFieldStringObject struct {
	Contents map[string]*FieldString
}

func NewLocalizedFieldStringObject() *LocalizedFieldStringObject {
	return &LocalizedFieldStringObject{
		Contents: make(map[string]*FieldString),
	}
}

func NewLocalizedFieldStringObjectWithContent(localization string, content *FieldString) *LocalizedFieldStringObject {
	obj := NewLocalizedFieldStringObject()
	obj.SetLocalizedContent(localization, content)
	return obj
}

func (obj *LocalizedFieldStringObject) SetLocalizedContent(localization string, content *FieldString) {
	// Don't overwrite existing content with empty content
	if existingContent, exists := obj.Contents[localization]; exists && content.IsEmpty() && !existingContent.IsEmpty() {
		return
	}
	obj.Contents[localization] = content
}

func (obj *LocalizedFieldStringObject) ReadAndSetLocalizedContent(localization string, bytes []byte, regularHeader, simplifiedHeader int) {
	if bytes == nil {
		return
	}

	charset := LocalizationToCharset(localization)
	fieldString := NewFieldString(charset, regularHeader, simplifiedHeader, bytes)
	obj.SetLocalizedContent(localization, fieldString)
}

func (obj *LocalizedFieldStringObject) WriteAllContent() string {
	var result []string

	for locKey, locName := range common.SupportedLanguages {
		if content, exists := obj.Contents[locKey]; exists && content != nil {
			result = append(result, fmt.Sprintf("[%s] %s", locName, content.String()))
		}
	}

	return strings.Join(result, "\n")
}

func (obj *LocalizedFieldStringObject) GetLocalizedContent(localization string) *FieldString {
	return obj.Contents[localization]
}

func (obj *LocalizedFieldStringObject) GetLocalizedString(localization string) string {
	if content := obj.GetLocalizedContent(localization); content != nil {
		return content.String()
	}
	return ""
}

func (obj *LocalizedFieldStringObject) GetDefaultContent() *FieldString {
	return obj.GetLocalizedContent(common.DefaultLocalization)
}

func (obj *LocalizedFieldStringObject) CopyInto(other *LocalizedFieldStringObject) {
	for localization, content := range obj.Contents {
		other.SetLocalizedContent(localization, content)
	}
}

func (obj *LocalizedFieldStringObject) String() string {
	if defaultContent := obj.GetDefaultContent(); defaultContent != nil {
		return defaultContent.String()
	}
	return ""
}
