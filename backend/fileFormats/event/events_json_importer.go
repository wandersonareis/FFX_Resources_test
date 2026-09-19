package event

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/common"
	"ffxresources/backend/models"
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
	return importEventJsonFile(gameVersion, "")
}

func ImportEventDataFromJsonFile(gameVersion common.GameVersion, eventID string) error {
	return importEventJsonFile(gameVersion, eventID)
}

func importEventJsonFile(gameVersion common.GameVersion, eventID string) error {
	jsonFilePath, err := getEventsJsonFilePath(gameVersion)
	if err != nil {
		return err
	}

	eventDataMap, err := loadEventJsonData(jsonFilePath)
	if err != nil {
		return err
	}

	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return processAllEventData(gameVersion, eventDataMap)
	}

	eventData, ok := eventDataMap[eventID]
	if !ok {
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	return processSingleEventData(gameVersion, eventData)
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

func processSingleEventData(gameVersion common.GameVersion, eventData EventFileData) error {
	common.LogVerbose("Looking for specific event: %s", eventData.ID)

	common.LogVerbose("Event %s found in JSON with %d strings", eventData.ID, len(eventData.Strings))

	if err := updateEventFromJsonData(gameVersion, eventData); err != nil {
		return fmt.Errorf("failed to update event %s: %w", eventData.ID, err)
	}

	if err := ExportEventStringsToLocalizations(gameVersion, eventData.ID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventData.ID, err)
		return fmt.Errorf("failed to save event %s: %w", eventData.ID, err)
	}

	common.LogVerbose("Event %s processed and saved successfully", eventData.ID)
	return nil
}

func processAllEventData(gameVersion common.GameVersion, eventDataMap map[string]EventFileData) error {
	common.LogVerbose("Processing all events from JSON (%d total)", len(eventDataMap))

	type eventResult struct {
		eventID string
		err     error
	}

	results := make(chan eventResult, len(eventDataMap))
	sem := make(chan struct{}, common.GetNumCpu())
	var wg sync.WaitGroup

	for eventID, eventData := range eventDataMap {
		wg.Add(1)
		go func(id string, data EventFileData) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			err := updateEventFromJsonData(gameVersion, data)
			results <- eventResult{eventID: id, err: err}
		}(eventID, eventData)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var failed []string
	for result := range results {
		if result.err != nil {
			common.LogError("failed to update event %s: %v", result.eventID, result.err)
			failed = append(failed, result.eventID)
		}
	}

	if len(failed) > 0 {
		sort.Strings(failed)
		return fmt.Errorf("failed to update %d event(s): %v", len(failed), failed)
	}

	common.LogVerbose("Events processed successfully! (%d events)", len(eventDataMap))
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

	if fieldString.GetRegularString() == newString {
		return nil
	}

	fieldString.SetRegularString(newString)
	return nil
}
