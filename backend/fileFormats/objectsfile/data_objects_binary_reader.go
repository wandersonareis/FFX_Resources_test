package objectsfile

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"fmt"
	"os"
	"path/filepath"
)

// ReadNameOnlyDataObjectsWithIlist reads binary data from a specified pattern path
// and creates NameOnlyDataObject entries with all available localizations.
//
// This function reads LocalizedTextObject entries containing only name information
// for various game elements that don't require descriptions. Each entry includes
// localized text for all supported languages in the game, making it suitable for
// simple text elements like labels, button text, or single-word entries.
//
// File format: name only data (binary format)
// Pattern path: Variable, passed as parameter (e.g., "battle/kernel/btl_txt.bin")
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: IList[ILocalizedTextObject] containing NameOnlyDataObject entries with full localization data
func ReadNameOnlyDataObjectsWithIlist(patternPath string) components.IList[datastore.IGlobalLocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
		return NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
	}

	nameObjects := ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)
	if nameObjects == nil || nameObjects.IsEmpty() {
		common.LogVerbose("No name-only data objects found for %s\n", patternPath)
		return nil
	}

	PopulateDataObjectLocalizationsWithIlist(patternPath, nameObjects, creator)

	common.LogVerbose("Loading %d name-only data objects...\n", nameObjects.Len())

	return nameObjects
}

// ReadNameDescriptionObjectsWithIlist reads binary data from a specified pattern path
// and creates NameDescriptionTextObject entries with all available localizations.
//
// This function reads LocalizedTextObject entries containing both name and description information
// for various game elements that require detailed text. Each entry includes localized text for all
// supported languages in the game, making it suitable for complex game objects like items, commands,
// abilities, and other elements that need both a title and detailed description.
//
// File format: name and description data (binary format)
// Pattern path: Variable, passed as parameter (e.g., "battle/kernel/command.bin")
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: IList[ILocalizedTextObject] containing NameDescriptionTextObject entries with full localization data
func ReadNameDescriptionObjectsWithIlist(patternPath string) components.IList[datastore.IGlobalLocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
		return NewNameDescriptionTextObject(data, stringBytes, headerLength, localization)
	}

	nameDescObjects := ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)
	if nameDescObjects == nil || nameDescObjects.IsEmpty() {
		common.LogVerbose("No name and description objects found for %s\n", patternPath)
		return nil
	}

	PopulateDataObjectLocalizationsWithIlist(patternPath, nameDescObjects, creator)

	common.LogVerbose("Loading %d name and description objects...\n", nameDescObjects.Len())

	return nameDescObjects
}

// PopulateDataObjectLocalizationsWithIlist populates localization data for all supported languages
// by reading corresponding binary files for each language and merging the localized content
// into the existing LocalizedTextObject instances within the provided IList.
//
// This function iterates through all supported languages, reads the binary files for each
// language, and applies the localized text content to the corresponding objects in the IList,
// leveraging IList's content() method for efficient access to the underlying collection.
//
// Parameters:
//   - path: Relative path to the binary file pattern (e.g., "battle/kernel/command.bin")
//   - objects: IList[ILocalizedTextObject] containing instances to be populated with localizations
//   - creator: Function that creates LocalizedTextObject instances from binary data
func PopulateDataObjectLocalizationsWithIlist(path string, objects components.IList[datastore.IGlobalLocalizedTextObject], creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) {
	if objects == nil || objects.IsEmpty() {
		return
	}

	for locKey := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRoot(locKey), path)

		localizationData := ReadDataListWithIlist(fullPath, locKey, creator)
		if localizationData != nil {
			maxLen := min(localizationData.Len(), objects.Len())
			items := objects.Items()
			locItems := localizationData.Items()
			for i := range maxLen {
				items[i].SetLocalizations(locItems[i])
			}
		}
	}
}

// ReadDataListWithIlist reads binary data files and creates localized text objects using a custom creator function.
//
// This is a high-level function that handles file reading and delegates to ParseDataListWithIlist for data processing.
// It provides convenient access to binary data files by handling file system operations and error checking,
// then passes the raw binary data to the parsing function for object creation.
//
// Parameters:
//   - filename: Path to the binary file to read
//   - languageCode: Language code for localization (e.g., "us", "jp", "de")
//   - creator: Function that creates ILocalizedTextObject instances from binary data
//     Parameters: (objData []byte, stringBytes []byte, headerLength int, languageCode string)
//
// Returns: IList[ILocalizedTextObject] containing localized text entries, or nil if reading fails
func ReadDataListWithIlist(filename string, languageCode string, creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) components.IList[datastore.IGlobalLocalizedTextObject] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return nil
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", filename)
		return nil
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return nil
	}

	return ParseDataListWithIlist(data, languageCode, creator)
}

// ParseDataListWithIlist parses binary data and creates localized text objects using a custom creator function.
//
// This is a low-level function that provides flexible binary data parsing for any type of localized text object.
// It handles the standard binary format used by the game's localization files, which includes a header section
// with object metadata and data sections containing object-specific information and string data.
//
// The function parses the binary data structure:
// - 8 bytes: header information
// - 2 bytes: minimum index value
// - 2 bytes: maximum index value
// - 2 bytes: individual object length
// - 2 bytes: total data section length
// - 4 bytes: reserved/padding
// - Variable: object data section
// - Variable: string data section
//
// Parameters:
//   - data: Binary data to parse
//   - languageCode: Language code for localization (e.g., "us", "jp", "de")
//   - creator: Function that creates ILocalizedTextObject instances from binary data
//     Parameters: (objData []byte, stringBytes []byte, headerLength int, languageCode string)
//
// Returns: IList[ILocalizedTextObject] containing localized text entries, or nil if parsing fails
func ParseDataListWithIlist(data []byte, languageCode string, creator func([]byte, []byte, int, string) datastore.IGlobalLocalizedTextObject) components.IList[datastore.IGlobalLocalizedTextObject] {
	if len(data) < 16 { // Minimum header size
		common.LogVerbose("Data too small for valid binary format")
		return nil
	}

	offset := 8

	minIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	maxIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Read individual length and total length
	individualLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	totalLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Skip 4 bytes
	offset += 4

	// Verify we have enough data
	if offset+totalLength > len(data) {
		common.LogVerbose("Insufficient data for specified total length")
		return nil
	}

	// Read data bytes
	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	// Read string bytes (remaining data)
	stringBytes := data[offset:]

	// Create IList with pre-calculated capacity for optimal memory allocation
	count := maxIndex - minIndex
	objectCount := count + 1
	objects := components.NewList[datastore.IGlobalLocalizedTextObject](objectCount)

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj := creator(objData, stringBytes, individualLength, languageCode)
		objects.Add(obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", (i*individualLength)+0x14)
			indexStr := fmt.Sprintf("%d", i+minIndex)
			objectString := obj.ToString(languageCode)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, objectString)
		}
	}

	return objects
}
