package reader

import (
	"bytes"
	"ffxresources/backend/builders"
	"ffxresources/backend/common"
	"ffxresources/backend/core/encoding"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/macrodic"
	"ffxresources/backend/formatters/json"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
)

func currentGameVersion() common.GameVersion {
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
		version = common.GameVersion(fieldStrings[0].Version)
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

	jsonFilePath := filepath.Join(jsonPath, common.WithVersionSuffixFor("events_all_localizations.json", currentGameVersion()))

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
	collection, err := json.NewJSONEventsFormatter().ReadEvents(jsonPath)
	if err != nil {
		return err
	}
	if err := builders.ApplyEventsDTO(currentGameVersion(), collection, nil); err != nil {
		return err
	}
	for eventID := range collection {
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
	jsonFilePath := filepath.Join(jsonPath, common.WithVersionSuffixFor("events_all_localizations.json", currentGameVersion()))

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Loading JSON file: %s", jsonFilePath)
	common.LogVerbose("Looking for event: %s", eventID)

	collection, err := json.NewJSONEventsFormatter().ReadEvents(jsonFilePath)
	if err != nil {
		return err
	}

	// O applier valida a existência no store; aqui só garantimos que o
	// evento pedido existe no DTO antes de aplicar.
	if _, ok := collection[eventID]; !ok {
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	common.LogVerbose("Event %s found in JSON with %d strings", eventID, len(collection[eventID].Rows))

	if err := builders.ApplyEventsDTO(currentGameVersion(), collection, []string{eventID}); err != nil {
		return err
	}

	// Write updated event back to files
	if err := ExportEventStringsToLocalizations(eventID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventID, err)
		return err
	}

	common.LogVerbose("Event %s processed and saved successfully", eventID)
	return nil
}

/*
JSON MACRO DICTIONARY EDITOR FUNCTIONS (DTO-based)
==================================================

This section imports the DTO-based JSON file created by ExportMacroDictionaryToJSON
(chunks chaveados por "chunk_NN" com rows de name/simplifiedName + hash) and
saves the rebuilt binaries back to the game files via builders.ApplyMacroDTO.
*/

// EditAndSaveMacroDictJSONFiles imports every localization in the macro
// dictionary JSON file and saves all rebuilt binaries.
func EditAndSaveMacroDictJSONFiles() error {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	jsonPath, err := json.DefaultMacroJSONPath(version)
	if err != nil {
		common.LogVerbose("Error resolving macro dictionary JSON file: %v", err)
		return err
	}
	collection, err := json.NewJSONMacroFormatter().ReadMacro(jsonPath)
	if err != nil {
		common.LogVerbose("Error loading macro dictionary JSON file: %v", err)
		return err
	}

	if err := builders.ApplyMacroDTO(version, collection); err != nil {
		common.LogVerbose("Error importing macro dictionary JSON: %v", err)
		return err
	}

	common.LogVerbose("Macro dictionary processed successfully")
	return nil
}

// EditAndSaveSpecificMacroDictFromJSON imports only the requested localization
// from the macro dictionary JSON file and saves its rebuilt binary.
//
// Parameters:
//   - localization: The localization code to process (e.g., "us", "jp", "de", etc.)
//
// Returns:
//   - error: nil if successful, error if the localization is not found or processing fails
func EditAndSaveSpecificMacroDictFromJSON(localization string) error {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	jsonPath, err := json.DefaultMacroJSONPath(version)
	if err != nil {
		common.LogVerbose("Error resolving macro dictionary JSON file: %v", err)
		return err
	}
	collection, err := json.NewJSONMacroFormatter().ReadMacro(jsonPath)
	if err != nil {
		common.LogVerbose("Error loading macro dictionary JSON file: %v", err)
		return err
	}

	containers, err := builders.RebuildMacroContainers(version, collection)
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
