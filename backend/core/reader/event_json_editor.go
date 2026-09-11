package reader

import (
	"bytes"
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
)

func currentGameVersion() models.GameVersion {
	return interactions.CurrentGameVersion()
}

func ExportEventStringsToLocalizations(eventID string) error {
	eventFile := event.GetEvent(currentGameVersion(), eventID)
	if eventFile == nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	if len(eventID) < 2 {
		return fmt.Errorf("invalid event ID: %s", eventID)
	}

	pathPattern := filepath.Join("event/obj_ps3/", eventID[:2], eventID, eventID+".bin")

	return writeStringsToStringToFileForAllLocalizations(pathPattern, eventFile.Strings)
}

func writeStringsToStringToFileForAllLocalizations(pathPattern string, localizedStrings []*event.LocalizedFieldStringObject) error {
	common.LogVerbose("Writing string file: %s", pathPattern)

	for localizationKey := range common.SupportedLanguages {
		localizationRoot := common.GetLocalizationRoot(localizationKey)

		localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, pathPattern)
		localePath = filepath.FromSlash(localePath)

		stringsBytes := stringsToStringFileBytes(localizedStrings, localizationKey)

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := common.WriteBytesToFile(localePath, stringsBytes); err != nil {
			return fmt.Errorf("failed to write file %s: %w", localePath, err)
		}
		common.LogVerbose("Wrote localized strings to %s (%d bytes)", localePath, len(stringsBytes))
	}
	return nil
}

func stringsToStringFileBytes(localizedStrings []*event.LocalizedFieldStringObject, languageCode string) []byte {
	if len(localizedStrings) == 0 {
		return []byte{}
	}

	charset := ffxencoding.GetCharsetForLanguage(languageCode)
	fieldStrings := make([]*event.FieldString, 0, len(localizedStrings))

	for _, localizedObj := range localizedStrings {
		if localizedObj == nil {
			continue
		}
		fieldString := localizedObj.GetLocalizedContent(languageCode)
		if fieldString == nil {
			fieldString = &event.FieldString{Charset: charset}
		}
		fieldStrings = append(fieldStrings, fieldString)
	}

	version := currentGameVersion()
	if len(fieldStrings) > 0 && fieldStrings[0] != nil {
		version = models.GameVersion(fieldStrings[0].Version)
	}
	stringBytes := event.RebuildFieldStrings(fieldStrings, charset, version)

	var buf bytes.Buffer
	for _, str := range fieldStrings {
		if str != nil {
			buf.Write(str.ToRegularHeaderBytes())
			buf.Write(str.ToSimplifiedHeaderBytes())
		}
	}
	buf.Write(stringBytes)

	return buf.Bytes()
}

func EditAndSaveEventJSONFiles() error {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")

	if !common.IsPathExists(jsonPath) {
		return fmt.Errorf("directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, common.WithVersionSuffix("events_all_localizations.json"))

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Processing JSON file: %s", jsonFilePath)

	if err := editAndSaveEventFromJSON(jsonFilePath); err != nil {
		common.LogVerbose("Error processing JSON file: %v", err)
		return err
	}
	common.LogVerbose("Events processed successfully!")
	return nil
}

func editAndSaveEventFromJSON(jsonPath string) error {
	loaded, loadErr := models.LoadDataFile[[]models.EventFileExport](jsonPath)
	var allEvents []EventFileData
	if loadErr != nil {
		raw, rawErr := common.ReadFile(jsonPath)
		if rawErr != nil {
			fmt.Printf("Erro ao ler arquivo JSON %s: %v\n", jsonPath, rawErr)
			return rawErr
		}
		if uErr := json.Unmarshal(raw, &allEvents); uErr != nil {
			fmt.Printf("Erro ao ler arquivo JSON %s: %v\n", jsonPath, uErr)
			return uErr
		}
	} else {
		allEvents = convertEventExports(loaded)
	}

	processedEventIDs := make(map[string]bool)

	for _, eventData := range allEvents {
		eventFile := event.GetEvent(currentGameVersion(), eventData.ID)
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
	}

	for eventID := range processedEventIDs {
		if err := ExportEventStringsToLocalizations(eventID); err != nil {
			return fmt.Errorf("error saving event %s: %w", eventID, err)
		}
	}

	return nil
}

// EditAndSaveSpecificEventFromJSON processes a specific event from the events_all_localizations.json file
// This function loads the JSON file, finds the specified event, and applies changes only to that event
//
// Parameters:
//   - eventID: The ID of the event to process (e.g., "ev001", "btl_001")
//
// Returns:
//   - error: nil if successful, error if the event is not found or processing fails
func EditAndSaveSpecificEventFromJSON(eventID string) error {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	jsonFilePath := filepath.Join(jsonPath, common.WithVersionSuffix("events_all_localizations.json"))

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Loading JSON file: %s", jsonFilePath)
	common.LogVerbose("Looking for event: %s", eventID)

	loaded, loadErr := models.LoadDataFile[[]models.EventFileExport](jsonFilePath)
	var allJsonEvents []EventFileData
	if loadErr != nil {
		raw, rawErr := common.ReadFile(jsonFilePath)
		if rawErr != nil {
			common.LogVerbose("Error opening JSON file %s: %v", jsonFilePath, rawErr)
			return rawErr
		}
		// Parse JSON content - expecting array of EventFileDataJSON
		if uErr := json.Unmarshal(raw, &allJsonEvents); uErr != nil {
			common.LogVerbose("Error decoding JSON %s: %v", jsonFilePath, uErr)
			return uErr
		}
	} else {
		allJsonEvents = convertEventExports(loaded)
	}

	// Find the specific event in the JSON
	var targetEventData *EventFileDataJSON
	for i := range allJsonEvents {
		if allJsonEvents[i].ID == eventID {
			targetEventData = &allJsonEvents[i]
			break
		}
	}

	// Check if event was found in JSON
	if targetEventData == nil {
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	common.LogVerbose("Event %s found in JSON with %d strings", eventID, len(targetEventData.Strings))

	// Validate event exists in memory
	eventFile := event.GetEvent(currentGameVersion(), eventID)
	if eventFile == nil {
		return fmt.Errorf("event not found in memory: %s", eventID)
	}

	common.LogVerbose("Processing event %s with %d strings", eventID, len(targetEventData.Strings))

	// Process each event string
	for _, eventString := range targetEventData.Strings {
		stringIndex := eventString.Index

		common.LogVerbose("Processing string %d for event %s", stringIndex, eventID)

		if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
			common.LogVerbose("String index out of range for event %s: %d", eventID, stringIndex)
			continue
		}

		objToEdit := eventFile.Strings[stringIndex]

		common.LogVerbose("Updating event %s[%d] with %d localizations",
			eventID, stringIndex, len(eventString.Text))

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

	// Write updated event back to files
	if err := ExportEventStringsToLocalizations(eventID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventID, err)
		return err
	}

	common.LogVerbose("Event %s processed and saved successfully", eventID)
	return nil
}

// EventFileDataJSON represents the JSON structure for a single event file
// This matches the format exported by WriteEventFileForAllLocalizationsJSON
type EventFileDataJSON = EventFileData

// convertEventExports maps the wrapped export payload back into the reader-internal
// EventFileData representation used by the JSON editors.
func convertEventExports(loaded []models.EventFileExport) []EventFileData {
	events := make([]EventFileData, 0, len(loaded))
	for _, e := range loaded {
		strings := make([]EventStringData, 0, len(e.Strings))
		for _, s := range e.Strings {
			strings = append(strings, EventStringData{Index: s.Index, Text: s.Text})
		}
		events = append(events, EventFileData{ID: e.ID, Strings: strings})
	}
	return events
}

// EventStringDataJSON represents a single event string with its localizations
// This matches the format exported by WriteEventFileForAllLocalizationsJSON
//type EventStringDataJSON = EventStringData

// EventFileData represents an event file with all its strings (same as writer package)
type EventFileData struct {
	ID      string            `json:"id"`
	Strings []EventStringData `json:"strings"`
}

// EventStringData represents a single event string with its localizations (same as writer package)
type EventStringData struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
}

/*
JSON MACRO DICTIONARY EDITOR FUNCTIONS (container-based)
=========================================================

This section imports the merged JSON file created by WriteMacroDictionaryJSON
(one entry per chunk/string holding every localization) and saves the rebuilt
binaries back to the game files.
*/

// EditAndSaveMacroDictJSONFiles imports every localization in the merged macro
// dictionary JSON file and saves all rebuilt binaries.
func EditAndSaveMacroDictJSONFiles() error {
	imp, err := macrodic.LoadMacroDictionaryJson(macrodic.MacroDictionaryJSONFileName)
	if err != nil {
		common.LogVerbose("Error loading macro dictionary JSON file: %v", err)
		return err
	}

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ImportFromJson(imp, version)
	if err != nil {
		common.LogVerbose("Error importing macro dictionary JSON: %v", err)
		return err
	}

	if err := macrodic.SaveMacroDictionaryBinaries(containers); err != nil {
		common.LogVerbose("Error saving macro dictionary binaries: %v", err)
		return err
	}

	common.LogVerbose("Macro dictionary processed successfully")
	return nil
}

// EditAndSaveSpecificMacroDictFromJSON imports only the requested localization
// from the merged macro dictionary JSON file and saves its rebuilt binary.
//
// Parameters:
//   - localization: The localization code to process (e.g., "us", "jp", "de", etc.)
//
// Returns:
//   - error: nil if successful, error if the localization is not found or processing fails
func EditAndSaveSpecificMacroDictFromJSON(localization string) error {
	imp, err := macrodic.LoadMacroDictionaryJson(macrodic.MacroDictionaryJSONFileName)
	if err != nil {
		common.LogVerbose("Error loading macro dictionary JSON file: %v", err)
		return err
	}

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	containers, err := macrodic.ImportFromJson(imp, version)
	if err != nil {
		common.LogVerbose("Error importing macro dictionary JSON: %v", err)
		return err
	}

	c, ok := containers[localization]
	if !ok || c == nil {
		return fmt.Errorf("localization %s not found in the JSON file", localization)
	}

	if err := macrodic.SaveMacroDictionaryBinaries(map[string]*macrodic.MacroDictionaryBinaryFile{localization: c}); err != nil {
		common.LogVerbose("Error saving macro dictionary binary for %s: %v", localization, err)
		return err
	}

	common.LogVerbose("Localization %s processed and saved successfully", localization)
	return nil
}

