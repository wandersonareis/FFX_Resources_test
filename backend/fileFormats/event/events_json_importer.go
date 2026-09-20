package event

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"ffxresources/backend/common"
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

func ImportEventsDataFromJsonFile(gameVersion common.GameVersion, formatter IEventsFormatter) error {
	return importEventJsonFile(gameVersion, "", formatter)
}

func ImportEventDataFromJsonFile(gameVersion common.GameVersion, eventID string, formatter IEventsFormatter) error {
	return importEventJsonFile(gameVersion, eventID, formatter)
}

func importEventJsonFile(gameVersion common.GameVersion, eventID string, formatter IEventsFormatter) error {
	jsonFilePath, err := getEventsJsonFilePath(gameVersion)
	if err != nil {
		return err
	}

	eventDataMap, err := loadEventJsonData(jsonFilePath, formatter)
	if err != nil {
		return err
	}

	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return processAllEventData(gameVersion, eventDataMap)
	}

	eventData, ok := findEventData(eventDataMap, eventID)
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

func loadEventJsonData(jsonFilePath string, formatter IEventsFormatter) ([]EventFileData, error) {
	common.LogVerbose("Loading events JSON file: %s", jsonFilePath)

	raw, err := common.ReadFile(jsonFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load events JSON file: %w", err)
	}
	events, err := formatter.Unmarshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to load events JSON file: %w", err)
	}

	common.LogVerbose("Successfully loaded %d events from JSON", len(events))
	return events, nil
}

// findEventData localiza um evento pelo ID no array importado (acesso por
// chave: a ordem não importa no import).
func findEventData(events []EventFileData, eventID string) (EventFileData, bool) {
	for _, e := range events {
		if e.ID == eventID {
			return e, true
		}
	}
	return EventFileData{}, false
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

func processAllEventData(gameVersion common.GameVersion, events []EventFileData) error {
	common.LogVerbose("Processing all events from JSON (%d total)", len(events))

	type eventResult struct {
		eventID string
		err     error
	}

	results := make(chan eventResult, len(events))
	sem := make(chan struct{}, common.GetNumCpu())
	var wg sync.WaitGroup

	for _, eventData := range events {
		wg.Add(1)
		go func(data EventFileData) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			err := updateEventFromJsonData(gameVersion, data)
			results <- eventResult{eventID: data.ID, err: err}
		}(eventData)
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

	common.LogVerbose("Events processed successfully! (%d events)", len(events))
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
