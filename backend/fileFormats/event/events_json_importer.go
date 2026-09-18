package event

import (
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
	"sort"
)

type (
	EventFileData struct {
		ID      string            `json:"id"`
		Strings []EventStringData `json:"strings"`
	}

	EventStringData struct {
		Index int               `json:"index"`
		Text  map[string]string `json:"text"`
	}
)

func ImportEventsDataFromJsonFile(gameVersion common.GameVersion) error {
	return importEventJsonFile(gameVersion, "", false)
}

func ImportEventDataFromJsonFile(gameVersion common.GameVersion, eventID string) error {
	return importEventJsonFile(gameVersion, eventID, true)
}

func importEventJsonFile(gameVersion common.GameVersion, eventID string, singleEvent bool) error {
	jsonFilePath, err := getEventsJsonFilePath(gameVersion)
	if err != nil {
		return err
	}

	eventDataMap, err := loadEventJsonData(jsonFilePath)
	if err != nil {
		return err
	}

	if singleEvent {
		return processSingleEventData(gameVersion, eventDataMap, eventID)
	}

	return processAllEventData(gameVersion, eventDataMap)
}

func getEventsJsonFilePath(gameVersion common.GameVersion) (string, error) {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if !common.IsPathExists(jsonPath) {
		return "", fmt.Errorf("edits directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, common.WithVersionSuffixFor("events_all_localizations.json", gameVersion))
	if !common.IsPathExists(jsonFilePath) {
		return "", fmt.Errorf("events JSON file not found: %s", jsonFilePath)
	}

	return jsonFilePath, nil
}

func loadEventJsonData(jsonFilePath string) (map[string]EventFileData, error) {
	common.LogVerbose("Loading events JSON file: %s", jsonFilePath)

	loaded, err := models.LoadDataFile[models.EventsFileExport](jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load events JSON file: %w", err)
	}

	eventDataMap := make(map[string]EventFileData, len(loaded.Strings))
	for eventID, export := range loaded.Strings {
		strings := make([]EventStringData, 0, len(export.Strings))
		for _, s := range export.Strings {
			strings = append(strings, EventStringData{Index: s.Index, Text: s.Text})
		}
		eventDataMap[eventID] = EventFileData{ID: eventID, Strings: strings}
	}

	common.LogVerbose("Successfully loaded %d events from JSON", len(eventDataMap))
	return eventDataMap, nil
}

func processSingleEventData(gameVersion common.GameVersion, eventDataMap map[string]EventFileData, eventID string) error {
	common.LogVerbose("Looking for specific event: %s", eventID)

	eventData, ok := eventDataMap[eventID]
	if !ok {
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	common.LogVerbose("Event %s found in JSON with %d strings", eventID, len(eventData.Strings))

	if err := updateEventFromJsonData(gameVersion, eventData); err != nil {
		return fmt.Errorf("failed to update event %s: %w", eventID, err)
	}

	if err := ExportEventStringsToLocalizations(gameVersion, eventID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventID, err)
		return fmt.Errorf("failed to save event %s: %w", eventID, err)
	}

	common.LogVerbose("Event %s processed and saved successfully", eventID)
	return nil
}

func processAllEventData(gameVersion common.GameVersion, eventDataMap map[string]EventFileData) error {
	common.LogVerbose("Processing all events from JSON (%d total)", len(eventDataMap))

	processedEventIDs := make(map[string]bool)

	keys := make([]string, 0, len(eventDataMap))
	for k := range eventDataMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, id := range keys {
		eventData := eventDataMap[id]
		if err := updateEventFromJsonData(gameVersion, eventData); err != nil {
			common.LogVerbose("failed to update event %s: %v", eventData.ID, err)
			continue
		}
		processedEventIDs[eventData.ID] = true
	}

	common.LogVerbose("Events processed successfully! (%d events)", len(processedEventIDs))
	return nil
}

func updateEventFromJsonData(gameVersion common.GameVersion, eventData EventFileData) error {
	eventFile := GetEvent(gameVersion, eventData.ID)
	if eventFile == nil {
		return fmt.Errorf("event not found in memory: %s", eventData.ID)
	}

	common.LogVerbose("Processing event %s with %d strings", eventData.ID, len(eventData.Strings))

	for _, eventString := range eventData.Strings {
		if err := updateEventStringFromJson(eventFile, eventString, eventData.ID); err != nil {
			common.LogVerbose("failed to update event %s: %v", eventData.ID, err)
			continue
		}
	}

	SetEvent(gameVersion, eventData.ID, eventFile)
	return nil
}

func updateEventStringFromJson(eventFile *EventFile, eventString EventStringData, eventID string) error {
	stringIndex := eventString.Index

	common.LogVerbose("Processing string %d for event %s", stringIndex, eventID)

	if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
		return fmt.Errorf("string index out of range for event %s: %d", eventID, stringIndex)
	}

	objToEdit := eventFile.Strings[stringIndex]

	common.LogVerbose("Updating event %s[%d] with %d localizations",
		eventID, stringIndex, len(eventString.Text))

	for localization, newString := range eventString.Text {
		if err := updateEventStringLocalization(objToEdit, localization, newString); err != nil {
			common.LogVerbose("failed to update localization %s for event %s[%d]: %v",
				localization, eventID, stringIndex, err)
		}
	}

	return nil
}

func updateEventStringLocalization(objToEdit *LocalizedFieldStringObject, localization, newString string) error {
	if newString == "" {
		return nil
	}

	if _, exists := common.SupportedLanguages[localization]; !exists {
		return fmt.Errorf("unsupported localization: %s", localization)
	}

	fieldString := objToEdit.GetLocalizedContent(localization)
	if fieldString == nil {
		return fmt.Errorf("failed to get localized content for %s", localization)
	}

	fieldString.SetRegularString(newString)
	return nil
}
