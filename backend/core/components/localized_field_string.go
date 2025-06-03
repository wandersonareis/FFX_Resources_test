package components

import (
	"ffxresources/backend/common"
	"fmt"
	"sort"
	"strings"
)

// LocalizedFieldStringObject manages field strings for different localizations
type LocalizedFieldStringObject struct {
	Contents map[string]*FieldString
}

// NewLocalizedFieldStringObject creates a new LocalizedFieldStringObject
func NewLocalizedFieldStringObject() *LocalizedFieldStringObject {
	return &LocalizedFieldStringObject{
		Contents: make(map[string]*FieldString),
	}
}

// NewLocalizedFieldStringObjectWithContent creates a new LocalizedFieldStringObject with initial content
func NewLocalizedFieldStringObjectWithContent(localization string, content *FieldString) *LocalizedFieldStringObject {
	obj := NewLocalizedFieldStringObject()
	obj.SetLocalizedContent(localization, content)
	return obj
}

// SetLocalizedContent sets content for a specific localization
func (obj *LocalizedFieldStringObject) SetLocalizedContent(localization string, content *FieldString) {
	// Don't overwrite existing content with empty content
	if existingContent, exists := obj.Contents[localization]; exists && content.IsEmpty() && !existingContent.IsEmpty() {
		return
	}
	obj.Contents[localization] = content
}

// ReadAndSetLocalizedContent reads field string data from bytes and sets it for a localization
func (obj *LocalizedFieldStringObject) ReadAndSetLocalizedContent(localization string, bytes []byte, regularHeader, simplifiedHeader int) {
	if bytes == nil {
		return
	}

	charset := LocalizationToCharset(localization)
	fieldString := NewFieldString(charset, regularHeader, simplifiedHeader, bytes)
	obj.SetLocalizedContent(localization, fieldString)
}

// WriteAllContent returns a formatted string with all localized content
func (obj *LocalizedFieldStringObject) WriteAllContent() string {
	var result []string

	for locKey, locName := range common.Localizations {
		if content, exists := obj.Contents[locKey]; exists && content != nil {
			result = append(result, fmt.Sprintf("[%s] %s", locName, content.String()))
		}
	}

	return strings.Join(result, "\n")
}

// GetLocalizedContent returns the field string for a specific localization
func (obj *LocalizedFieldStringObject) GetLocalizedContent(localization string) *FieldString {
	return obj.Contents[localization]
}

// GetLocalizedString returns the string content for a specific localization
func (obj *LocalizedFieldStringObject) GetLocalizedString(localization string) string {
	if content := obj.GetLocalizedContent(localization); content != nil {
		return content.String()
	}
	return ""
}

// GetDefaultContent returns the content for the default localization
func (obj *LocalizedFieldStringObject) GetDefaultContent() *FieldString {
	// Assuming "us" is the default localization - this should match project constants
	return obj.GetLocalizedContent("us")
}

// CopyInto copies all content into another LocalizedFieldStringObject
func (obj *LocalizedFieldStringObject) CopyInto(other *LocalizedFieldStringObject) {
	for localization, content := range obj.Contents {
		other.SetLocalizedContent(localization, content)
	}
}

// WriteAllContentToCsv exports all localized content to CSV format
func (obj *LocalizedFieldStringObject) WriteAllContentToCsv() string {
	// Get sorted localization keys
	var localizationKeys []string
	for key := range common.Localizations {
		localizationKeys = append(localizationKeys, key)
	}
	sort.Strings(localizationKeys)

	var csvBuilder strings.Builder

	// Write header
	csvBuilder.WriteString("\"string index\"")
	for _, langKey := range localizationKeys {
		csvBuilder.WriteString(",\"")
		csvBuilder.WriteString(langKey)
		csvBuilder.WriteString("\"")
	}
	csvBuilder.WriteString("\n")

	// Write index value (using default content)
	defaultFieldString := obj.GetDefaultContent()
	indexValue := ""
	if defaultFieldString != nil {
		if tempVal := defaultFieldString.String(); tempVal != "" {
			indexValue = tempVal
		}
	}
	csvBuilder.WriteString("\"")
	csvBuilder.WriteString(obj.escapeCsvValue(indexValue))
	csvBuilder.WriteString("\"")

	// Write localized values
	for _, langKey := range localizationKeys {
		localizedString := obj.GetLocalizedContent(langKey)
		cellValue := ""
		if localizedString != nil {
			cellValue = localizedString.String()
		}
		csvBuilder.WriteString(",\"")
		csvBuilder.WriteString(obj.escapeCsvValue(cellValue))
		csvBuilder.WriteString("\"")
	}

	return csvBuilder.String()
}

// escapeCsvValue escapes quotes in CSV values
func (obj *LocalizedFieldStringObject) escapeCsvValue(value string) string {
	return strings.ReplaceAll(value, "\"", "\"\"")
}

// String returns the string representation using default content
func (obj *LocalizedFieldStringObject) String() string {
	if defaultContent := obj.GetDefaultContent(); defaultContent != nil {
		return defaultContent.String()
	}
	return ""
}
