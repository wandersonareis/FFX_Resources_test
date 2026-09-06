package event

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"fmt"
	"path/filepath"
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

	charset := ffxencoding.GetCharsetForLanguage(localization)
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

// ReadStringFile reads string file(s) from the given filename path
// If the path is a directory, it recursively reads all files within it
// Returns a slice of FieldString objects parsed from the file data
//
// Parameters:
//   - filename: Path to file or directory to read
//   - localization: Localization code (e.g., "jp", "us", "kr") for charset conversion
//
// Returns:
//   - []*FieldString: Slice of parsed FieldString objects, or nil if directory or error
//
// Behavior:
//   - For directories: Recursively processes all non-hidden files in sorted order
//   - For files: Resolves path, reads bytes, and parses as string data using appropriate charset
func ReadStringFile(filename string, languageCode string) []*FieldString {
	resolvedPath, err := common.NewFileAccessor(filename)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Error resolving file %s: %v\n", filename, err)
		}
		return nil
	}

	bytes := resolvedPath.ReadBytes()
	if bytes == nil {
		if common.IsVerboseMode() {
			fmt.Printf("Failed to read bytes from file %s\n", resolvedPath.ResolvedPath)
		}
		return nil
	}

	charset := ffxencoding.GetCharsetForLanguage(languageCode)
	fieldStrings, err := FromFieldStringData(bytes, charset)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Error parsing string data from %s: %v\n", filename, err)
		}
		return nil
	}
	return fieldStrings
}

// ReadLocalizedStringFiles reads localized string files for all available localizations
// This function iterates through all localizations and reads string files for each one
//
// Parameters:
//   - path: Relative path to the string file (e.g., "event/obj_ps3/XX/XXXX/XXXX.bin")
//
// Returns:
//   - []*LocalizedFieldStringObject: Slice of localized string objects with content for each localization
//
// Behavior:
//   - Iterates through all localizations defined in common.Localizations
//   - For each localization, constructs full path using GetLocalizationRoot + path
//   - Reads string files using ReadStringFile
//   - Merges all localized content into LocalizedFieldStringObject instances
//   - Each index in the returned slice contains all localizations for that string
func ReadLocalizedStringFiles(path string) []*LocalizedFieldStringObject {
	localized := make([]*LocalizedFieldStringObject, 0)

	for key := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRoot(key), path)
		localizedStrings := ReadStringFile(fullPath, key)

		for i, fieldString := range localizedStrings {
			for len(localized) <= i {
				localized = append(localized, NewLocalizedFieldStringObject())
			}

			localized[i].SetLocalizedContent(key, fieldString)
		}
	}

	return localized
}

func ReadLocalizedEventStrings(eventId string) ([]*LocalizedFieldStringObject, error) {
	if len(eventId) < 2 {
		return nil, fmt.Errorf("invalid event ID: %s", eventId)
	}
	shortened := eventId[:2]
	midPath := filepath.Join(shortened, eventId, eventId)
	localizedStrings := ReadLocalizedStringFiles("event/obj_ps3/" + midPath + ".bin")
	if localizedStrings == nil {
		return nil, fmt.Errorf("failed to read localized strings for event %s", eventId)
	}

	return localizedStrings, nil
}
