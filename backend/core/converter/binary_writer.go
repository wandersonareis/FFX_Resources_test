package converter

import (
	"bytes"
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
)

// ExportLocalizedTextData exports localized text data from an IList to binary files
// for all supported localizations. This function converts in-memory localized objects
// back to their original binary format and writes them to the appropriate files.
//
// The function validates that objects are loaded, converts the IList to a slice,
// and delegates to the lower-level writing functions to handle the actual file operations.
//
// Parameters:
//   - objects: IList containing ILocalizedTextObject entries to be exported
//   - pathPattern: Relative path pattern for the binary files (e.g., "battle/kernel/command.bin")
//
// Returns: error if no objects are loaded or if the write operation fails
func ExportLocalizedTextData(objects components.IList[components.ILocalizedTextObject], pathPattern string) error {
	if objects.IsEmpty() {
		return fmt.Errorf("nenhum objeto carregado")
	}

	dataObjects := make([]components.ILocalizedTextObject, objects.GetLength())
	for i, obj := range objects.GetItems() {
		dataObjects[i] = obj
	}

	from := 0
	to := len(dataObjects)

	return writeLocalizedDataObjectsInAllLocalizations(pathPattern, dataObjects, from, to)
}

// writeLocalizedDataObjectsInAllLocalizations writes localized data objects to binary files
// for all supported localizations. This function handles the core logic of converting
// in-memory objects back to binary format and writing them to the appropriate file paths.
//
// The function iterates through all supported languages, converts objects to bytes for each
// localization, ensures directory structure exists, and writes the binary data to files.
//
// Parameters:
//   - pathPattern: Relative path pattern for the binary files
//   - objects: Slice of ILocalizedTextObject instances to be written
//   - startIndex: Starting index in the objects slice (inclusive)
//   - endIndex: Ending index in the objects slice (exclusive)
//
// Returns: error if indices are invalid or if any write operation fails
func writeLocalizedDataObjectsInAllLocalizations(pathPattern string, objects []components.ILocalizedTextObject, startIndex, endIndex int) error {
	common.LogVerbose("Writing localized data objects to: %s (from %d to %d)", pathPattern, startIndex, endIndex)

	if startIndex < 0 || endIndex > len(objects) || startIndex >= endIndex {
		return fmt.Errorf("invalid indices: from=%d, to=%d, length=%d", startIndex, endIndex, len(objects))
	}

	for localizationKey := range common.SupportedLanguages {
		if localizationKey != "us" {
			continue // Skip non-US localizations for now
		}
		localizationRoot := common.GetLocalizationRoot(localizationKey)
		localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, pathPattern)
		localePath = filepath.FromSlash(localePath)

		dataBytes, err := convertFFXLocalizedDataToBytes(objects, startIndex, endIndex, localizationKey)
		if err != nil {
			return fmt.Errorf("error when converting objects to bytes (localization %s): %w", localizationKey, err)
		}

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("error when creating directory %s: %w", dir, err)
		}

		if err := common.WriteBytesToFile(localePath, dataBytes); err != nil {
			return fmt.Errorf("error when writing file %s: %w", localePath, err)
		}

		common.LogVerbose("Wrote localized data to %s (%d bytes)", localePath, len(dataBytes))
	}

	return nil
}

// convertFFXLocalizedDataToBytes converts a slice of ILocalizedTextObject instances to binary format
// for a specific localization. This function recreates the original binary file structure including
// headers, object data, and string data sections.
//
// The binary format consists of:
// 1. FFX header (8 bytes)
// 2. Object range information (4 bytes)
// 3. Header length and total length (4 bytes)
// 4. Unknown header bytes (4 bytes)
// 5. Object data section (variable length)
// 6. String data section (variable length)
//
// Parameters:
//   - objects: Slice of ILocalizedTextObject instances to convert
//   - from: Starting index in the objects slice
//   - to: Ending index in the objects slice
//   - languageCode: Language code for the target localization
//
// Returns: byte slice containing the binary data, or error if conversion fails
func convertFFXLocalizedDataToBytes(objects []components.ILocalizedTextObject, from, to int, languageCode string) ([]byte, error) {
	if len(objects) == 0 {
		return nil, fmt.Errorf("objects slice is empty")
	}

	objectHeaderLength := objects[0].GetHeaderLength()

	ffxHeader := []byte{1, 0, 0, 0, 0, 0, 0, 0}
	ffxUnknownBytesOfHeader := []byte{0x14, 0x00, 0x00, 0x00}

	var allKeyedStrings []*components.KeyedString
	for _, obj := range objects {
		keyedStrings := obj.GetLocalizedKeyedStrings(languageCode)
		for _, ks := range keyedStrings {
			if ks != nil {
				allKeyedStrings = append(allKeyedStrings, ks)
			} else {
				// TODO: delete this
				common.LogVerbose("Keyed string is nil for object at index %d", obj.GetName(common.DefaultLocalization))
			}
		}
	}

	charset := components.GetCharsetForLanguage(languageCode)
	stringBytes := components.RebuildKeyedStrings(allKeyedStrings, charset)

	var buf bytes.Buffer

	// Write FFX header
	buf.Write(ffxHeader)

	// Write object range (from index)
	binary.Write(&buf, binary.LittleEndian, uint16(from))

	// Write object range (to index - 1)
	toMinus1 := to - 1
	binary.Write(&buf, binary.LittleEndian, uint16(toMinus1))

	// Write header length per object
	binary.Write(&buf, binary.LittleEndian, uint16(objectHeaderLength))

	// Write total header section length
	totalLength := len(objects) * objectHeaderLength
	binary.Write(&buf, binary.LittleEndian, uint16(totalLength))

	// Write unknown header bytes
	buf.Write(ffxUnknownBytesOfHeader)

	// Write object data section
	for _, obj := range objects {
		buf.Write(obj.ToBytes(languageCode))
	}

	// Write string data section
	buf.Write(stringBytes)

	return buf.Bytes(), nil
}

// ExportEventsForLocalizationToJSON exports all events in EVENTS to binary files for
// all supported localizations. This function iterates through all loaded events
// and exports each one using ExportEventStringsToLocalizations.
//
// This function directly processes the EVENTS data structure in memory,
// bypassing any JSON file reading logic, making it more efficient and ensuring
// that the current state of EVENTS (updated or not) is exported.
//
// Returns: error if no events are loaded or if any export operation fails
func ExportAllEventsForLocalizations() error {
	if len(components.EVENTS) == 0 {
		return fmt.Errorf("no events loaded in EVENTS data structure")
	}

	common.LogVerbose("Exporting %d events from EVENTS data structure", len(components.EVENTS))

	var exportedCount int
	var failedEvents []string

	for eventID := range components.EVENTS {
		if err := ExportEventStringsToLocalizations(eventID); err != nil {
			common.LogVerbose("Failed to export event %s: %v", eventID, err)
			failedEvents = append(failedEvents, eventID)
			continue
		}

		exportedCount++
	}

	common.LogVerbose("Successfully exported %d events", exportedCount)

	if len(failedEvents) > 0 {
		common.LogVerbose("Failed to export %d events: %v", len(failedEvents), failedEvents)
		return fmt.Errorf("failed to export %d events: %v", len(failedEvents), failedEvents)
	}

	return nil
}

// ExportEventStringsToLocalizations exports event strings from the EVENTS data structure
// to binary files for all supported localizations. This function handles the conversion
// of event text data to the simplified binary format used by event files.
//
// Event files use a simpler binary format compared to other localized data:
// - No complex FFX header structure
// - Direct string data with simplified headers
// - String pointers/references only
//
// Parameters:
//   - eventID: The unique identifier for the event (e.g., "ev001", "btl_001")
//
// Returns: error if event is not found or export fails
func ExportEventStringsToLocalizations(eventID string) error {
	eventFile, exists := components.EVENTS[eventID]
	if !exists || eventFile == nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	if len(eventID) < 2 {
		return fmt.Errorf("invalid event ID: %s", eventID)
	}

	if len(eventFile.Strings) == 0 {
		return fmt.Errorf("no strings to write for event: %s", eventID)
	}

	pathPattern := filepath.Join("event/obj_ps3/", eventID[:2], eventID, eventID+".bin")

	return writeEventStringsToAllLocalizations(pathPattern, eventFile.Strings)
}

// writeEventStringsToAllLocalizations writes event string data to binary files for all
// supported localizations. This function handles the file I/O operations and delegates
// the binary conversion to specialized functions.
//
// The function iterates through all supported languages, converts the event strings
// to the appropriate binary format, ensures directory structure exists, and writes
// the data to the target files.
//
// Parameters:
//   - pathPattern: Relative path pattern for the event binary files
//   - localizedStrings: Slice of LocalizedFieldStringObject containing the event text data
//
// Returns: error if directory creation or file writing fails
func writeEventStringsToAllLocalizations(pathPattern string, localizedStrings []*components.LocalizedFieldStringObject) error {
	common.LogVerbose("Writing event strings to: %s", pathPattern)

	if len(localizedStrings) == 0 {
		common.LogVerbose("No event strings to write for path: %s", pathPattern)
		return nil
	}

	for localizationKey := range common.SupportedLanguages {
		localizationRoot := common.GetLocalizationRoot(localizationKey)
		localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, pathPattern)
		localePath = filepath.FromSlash(localePath)

		stringBytes, err := convertEventStringsToBytes(localizedStrings, localizationKey)
		if err != nil {
			return fmt.Errorf("error converting event strings to bytes (localization %s): %w", localizationKey, err)
		}

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}

		if err := common.WriteBytesToFile(localePath, stringBytes); err != nil {
			return fmt.Errorf("error writing event file %s: %w", localePath, err)
		}

		common.LogVerbose("Wrote event strings to %s (%d bytes)", localePath, len(stringBytes))
	}

	return nil
}

// convertEventStringsToBytes converts event string objects to binary format for a specific
// localization. This function handles the simplified binary format used by event files,
// which differs from the complex FFX format used by other game data.
//
// Event binary format structure:
// 1. Field string headers (simplified format)
// 2. String content data (encoded with appropriate charset)
//
// The format is simpler than other game data files:
// - No FFX header bytes
// - No object range information
// - Direct string data with minimal headers
//
// Parameters:
//   - localizedStrings: Slice of LocalizedFieldStringObject to convert
//   - languageCode: Target language code for the conversion
//
// Returns: byte slice containing the binary data, or error if conversion fails
func convertEventStringsToBytes(localizedStrings []*components.LocalizedFieldStringObject, languageCode string) ([]byte, error) {
	if len(localizedStrings) == 0 {
		return []byte{}, nil
	}

	charset := components.GetCharsetForLanguage(languageCode)
	fieldStrings := extractFieldStringsForLanguage(localizedStrings, languageCode, charset)

	return buildEventStringsBinaryData(fieldStrings)
}

// extractFieldStringsForLanguage extracts FieldString objects for a specific language
// from LocalizedFieldStringObject instances. This function handles the conversion from
// localized objects to the field string format used in binary files.
//
// Parameters:
//   - localizedStrings: Source localized string objects
//   - languageCode: Target language code
//   - charset: Character set to use for string encoding
//
// Returns: slice of FieldString objects ready for binary conversion
func extractFieldStringsForLanguage(localizedStrings []*components.LocalizedFieldStringObject, languageCode string, charset string) []*components.FieldString {
	fieldStrings := make([]*components.FieldString, 0, len(localizedStrings))

	for _, localizedObj := range localizedStrings {
		if localizedObj == nil {
			continue
		}

		fieldString := localizedObj.GetLocalizedContent(languageCode)
		if fieldString == nil {
			// Create empty field string with appropriate charset if content is missing
			fieldString = &components.FieldString{Charset: charset}
		}

		fieldStrings = append(fieldStrings, fieldString)
	}

	return fieldStrings
}

// buildEventStringsBinaryData constructs the final binary data for event strings.
// This function combines the string headers and content into the format expected
// by the game's event system.
//
// Binary structure:
// 1. Regular headers for each string
// 2. Simplified headers for each string
// 3. String content data (rebuilt from FieldString objects)
//
// Parameters:
//   - fieldStrings: FieldString objects to convert to binary format
//
// Returns: complete binary data ready for file writing, or error if rebuild fails
func buildEventStringsBinaryData(fieldStrings []*components.FieldString) ([]byte, error) {
	if len(fieldStrings) == 0 {
		return []byte{}, nil
	}

	// Rebuild the string content data using existing components functionality
	stringBytes := components.RebuildFieldStrings(fieldStrings, fieldStrings[0].Charset)

	var buf bytes.Buffer

	// Write headers for each field string
	for _, str := range fieldStrings {
		if str != nil {
			buf.Write(str.ToRegularHeaderBytes())
			buf.Write(str.ToSimplifiedHeaderBytes())
		}
	}

	// Append the string content data
	buf.Write(stringBytes)

	return buf.Bytes(), nil
}

// editAndSaveEventFromJSON processes a JSON file containing event data and applies
// changes to the EVENTS data structure. This function reads the JSON file, updates
// the in-memory event objects, and exports them back to binary files using the
// converter's binary writer functions.
//
// The function processes each event in the JSON file, updating the localized strings
// for each event, and then exports the updated events to binary files for all
// supported localizations.
//
// Parameters:
//   - jsonPath: Absolute path to the JSON file containing event data
//
// Returns: error if file reading, JSON parsing, or export fails
func editAndSaveEventFromJSON(jsonPath string) error {
	allEvents, err := common.ReadJsonFile[[]EventFileData](jsonPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo JSON %s: %w", jsonPath, err)
	}

	processedEventIDs := make(map[string]bool)

	for _, eventData := range allEvents {
		eventFile, exists := components.EVENTS[eventData.ID]
		if !exists || eventFile == nil {
			common.LogVerbose("Event not found: %s", eventData.ID)
			continue
		}

		processedEventIDs[eventData.ID] = true
		common.LogVerbose("Processing event %s with %d strings", eventData.ID, len(eventData.Strings))

		for _, eventString := range eventData.Strings {
			stringIndex := eventString.Index
			common.LogVerbose("Processing string %d for event %s", stringIndex, eventData.ID)

			if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
				common.LogVerbose("String index out of range for event %s: %d", eventData.ID, stringIndex)
				continue
			}

			objToEdit := eventFile.Strings[stringIndex]

			common.LogVerbose("Updating event %s[%d] with %d localizations",
				eventData.ID, stringIndex, len(eventString.Text))

			for localization, newString := range eventString.Text {
				if newString != "" {
					if _, exists := common.SupportedLanguages[localization]; exists {
						fieldString := objToEdit.GetLocalizedContent(localization)
						if fieldString != nil {
							fieldString.SetRegularString(newString)
						}
					}
				}
			}
		}
	}

	// Export all processed events to binary files
	for eventID := range processedEventIDs {
		if err := ExportEventStringsToLocalizations(eventID); err != nil {
			return fmt.Errorf("error saving event %s: %w", eventID, err)
		}
	}

	return nil
}
