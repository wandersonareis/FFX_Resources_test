package event

import (
	"bytes"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
)

// ExportAllEventsForLocalizations exports all events in EVENTS to binary files for
// all supported localizations. This function iterates through all loaded events
// and exports each one using ExportEventStringsToLocalizations.
//
// This function directly processes the events from the datastore,
// bypassing any JSON file reading logic, making it more efficient and ensuring
// that the current state of events (updated or not) is exported.
//
// Returns: error if no events are loaded or if any export operation fails
func ExportAllEventsForLocalizations() error {
	eventIDs := GetAllEventIDs()
	if len(eventIDs) == 0 {
		return fmt.Errorf("no events loaded in datastore")
	}

	common.LogVerbose("Exporting %d events from datastore", len(eventIDs))

	var exportedCount int
	var failedEvents []string

	for _, eventID := range eventIDs {
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

// ExportEventStringsToLocalizations exports event strings from the datastore
// to binary files for all supported localizations. This function handles the conversion
// of event text data to the simplified binary format used by event files.
//
// Event files use a simpler binary format compared to other localized data:
// - No complex FFX objects data header structure
// - Direct string data with simplified headers
// - String pointers/references only
//
// Parameters:
//   - eventID: The unique identifier for the event (e.g., "ev001", "btl_001")
//
// Returns: error if event is not found or export fails
func ExportEventStringsToLocalizations(eventID string) error {
	eventFile := GetEvent(eventID)
	if eventFile == nil {
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
func writeEventStringsToAllLocalizations(pathPattern string, localizedStrings []*LocalizedFieldStringObject) error {
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
func convertEventStringsToBytes(localizedStrings []*LocalizedFieldStringObject, languageCode string) ([]byte, error) {
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
func extractFieldStringsForLanguage(localizedStrings []*LocalizedFieldStringObject, languageCode string, charset string) []*FieldString {
	fieldStrings := make([]*FieldString, 0, len(localizedStrings))

	for _, localizedObj := range localizedStrings {
		if localizedObj == nil {
			continue
		}

		fieldString := localizedObj.GetLocalizedContent(languageCode)
		if fieldString == nil {
			fieldString = NewEmptyFieldString(charset)
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
func buildEventStringsBinaryData(fieldStrings []*FieldString) ([]byte, error) {
	if len(fieldStrings) == 0 {
		return []byte{}, nil
	}

	stringBytes := RebuildFieldStrings(fieldStrings, fieldStrings[0].Charset)

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

// EditAndSaveEventFromJSON processes a JSON file containing event data and applies
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
func EditAndSaveEventFromJSON(jsonPath string) error {
	allEvents, err := common.ReadJsonFile[[]EventFileData](jsonPath)
	if err != nil {
		return fmt.Errorf("error reading json file %s: %w", jsonPath, err)
	}

	processedEventIDs := make(map[string]bool)

	for _, eventData := range allEvents {
		eventFile := GetEvent(eventData.ID)
		if eventFile == nil {
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

		// Atualizar o evento no datastore após modificações
		SetEvent(eventData.ID, eventFile)
	}

	for eventID := range processedEventIDs {
		if err := ExportEventStringsToLocalizations(eventID); err != nil {
			return fmt.Errorf("error saving event %s: %w", eventID, err)
		}
	}

	return nil
}
