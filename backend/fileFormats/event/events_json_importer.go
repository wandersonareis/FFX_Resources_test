package event

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
)

type (
	EventFileData struct {
		ID      string            `json:"id"`
		Strings []EventStringData `json:"strings"`
	}

	// EventStringData represents a single event string with its localizations for JSON processing
	EventStringData struct {
		Index int               `json:"index"`
		Text  map[string]string `json:"text"`
	}
)

// ImportEventsDataFromJsonFile imports event data from a JSON file and applies
// the translations to all events found in the JSON file.
//
// This function reads JSON files containing event string information for all
// supported languages and applies the translations directly to the global EVENTS
// variable. Each event contains multiple strings with localized text content.
//
// File format: JSON array with event id and strings array containing localizations
// JSON structure: [{"id": "ev001", "strings": [{"index": 0, "text": {"us": "...", "sp": "..."}}]}]
//
// JSON file: events_all_localizations.json
// Target: EVENTS (multiple entries)
//
// Returns: error if import fails or file cannot be read
func ImportEventsDataFromJsonFile(gameVersion models.GameVersion) error {
	return importEventJsonFile(gameVersion, "", false)
}

// ImportEventDataFromJsonFile imports event data from a JSON file and applies
// the translations to a single specified event.
//
// This function reads JSON files containing event string information and applies
// the translations to a specific event in the EVENTS variable. The event
// is identified by its ID and must exist in the JSON file.
//
// JSON file: events_all_localizations.json
// Target: EVENTS (single entry specified by eventID)
//
// Parameters:
//   - eventID: The ID of the specific event to process (e.g., "ev001", "btl_001")
//
// Returns: error if import fails, file cannot be read, or event is not found
func ImportEventDataFromJsonFile(gameVersion models.GameVersion, eventID string) error {
	return importEventJsonFile(gameVersion, eventID, true)
}

// importEventJsonFile is the core function that handles both single and multiple event processing.
// This function centralizes the common logic between ImportEventsDataFromJsonFile and ImportEventDataFromJsonFile
// to avoid code duplication while providing flexibility for different processing modes.
//
// Parameters:
//   - eventID: The specific event ID to process (empty string for all events)
//   - singleEvent: Whether to process only a single event (true) or all events (false)
//
// Returns: error if processing fails
func importEventJsonFile(gameVersion models.GameVersion, eventID string, singleEvent bool) error {
	jsonFilePath, err := getEventsJsonFilePath()
	if err != nil {
		return err
	}

	eventDataList, err := loadEventJsonData(jsonFilePath)
	if err != nil {
		return err
	}

	if singleEvent {
		return processSingleEventData(gameVersion, eventDataList, eventID)
	}

	return processAllEventData(gameVersion, eventDataList)
}

// getEventsJsonFilePath constructs and validates the path to the events JSON file.
// This function handles the standard path construction and existence validation.
//
// Returns: validated file path string, or error if file doesn't exist
func getEventsJsonFilePath() (string, error) {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if !common.IsPathExists(jsonPath) {
		return "", fmt.Errorf("edits directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, common.WithVersionSuffix("events_all_localizations.json"))
	if !common.IsPathExists(jsonFilePath) {
		return "", fmt.Errorf("events JSON file not found: %s", jsonFilePath)
	}

	return jsonFilePath, nil
}

// loadEventJsonData loads and parses the events JSON file into EventFileData structures.
// This function handles file reading, JSON parsing, and basic validation.
//
// Parameters:
//   - jsonFilePath: Path to the events JSON file
//
// Returns: slice of EventFileData structures, or error if loading/parsing fails
func loadEventJsonData(jsonFilePath string) ([]EventFileData, error) {
	common.LogVerbose("Loading events JSON file: %s", jsonFilePath)

	loaded, loadErr := models.LoadDataFile[[]models.EventFileExport](jsonFilePath)
	var eventDataList []EventFileData
	if loadErr != nil {
		// Fallback to a raw JSON array (legacy format without wrapper).
		raw, rawErr := common.ReadFile(jsonFilePath)
		if rawErr != nil {
			common.LogVerbose("Error reading events JSON file %s: %v", jsonFilePath, rawErr)
			return nil, fmt.Errorf("failed to read events JSON file: %w", rawErr)
		}
		if err := json.Unmarshal(raw, &eventDataList); err != nil {
			common.LogVerbose("Error parsing events JSON file %s: %v", jsonFilePath, err)
			return nil, fmt.Errorf("failed to parse events JSON file: %w", err)
		}
	} else {
		eventDataList = make([]EventFileData, 0, len(loaded))
		for _, e := range loaded {
			strings := make([]EventStringData, 0, len(e.Strings))
			for _, s := range e.Strings {
				strings = append(strings, EventStringData{Index: s.Index, Text: s.Text})
			}
			eventDataList = append(eventDataList, EventFileData{ID: e.ID, Strings: strings})
		}
	}

	common.LogVerbose("Successfully loaded %d events from JSON", len(eventDataList))
	return eventDataList, nil
}

// processSingleEventData processes a single event from the JSON data.
// This function finds the specified event in the data and applies its changes.
//
// Parameters:
//   - eventDataList: List of all event data from JSON
//   - eventID: The specific event ID to process
//
// Returns: error if event is not found or processing fails
func processSingleEventData(gameVersion models.GameVersion, eventDataList []EventFileData, eventID string) error {
	common.LogVerbose("Looking for specific event: %s", eventID)

	var targetEventData *EventFileData
	for i := range eventDataList {
		if eventDataList[i].ID == eventID {
			targetEventData = &eventDataList[i]
			break
		}
	}

	if targetEventData == nil {
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	common.LogVerbose("Event %s found in JSON with %d strings", eventID, len(targetEventData.Strings))

	if err := updateEventFromJsonData(gameVersion, *targetEventData); err != nil {
		return fmt.Errorf("failed to update event %s: %w", eventID, err)
	}

	if err := ExportEventStringsToLocalizations(gameVersion, eventID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventID, err)
		return fmt.Errorf("failed to save event %s: %w", eventID, err)
	}

	common.LogVerbose("Event %s processed and saved successfully", eventID)
	return nil
}

// processAllEventData processes all events from the JSON data.
// This function iterates through all events and applies their changes.
//
// Parameters:
//   - eventDataList: List of all event data from JSON
//
// Returns: error if any critical processing fails
func processAllEventData(gameVersion models.GameVersion, eventDataList []EventFileData) error {
	common.LogVerbose("Processing all events from JSON (%d total)", len(eventDataList))

	processedEventIDs := make(map[string]bool)

	for _, eventData := range eventDataList {
		if err := updateEventFromJsonData(gameVersion, eventData); err != nil {
			common.LogVerbose("failed to update event %s: %v", eventData.ID, err)
			continue
		}
		processedEventIDs[eventData.ID] = true
	}

	common.LogVerbose("Events processed successfully! (%d events)", len(processedEventIDs))
	return nil
}

// updateEventFromJsonData updates a single event's data based on JSON input.
// This function handles the core logic of applying JSON changes to event strings.
// Uses datastore as the single source of truth for event data.
//
// Parameters:
//   - eventData: The event data from JSON containing updates to apply
//
// Returns: error if the event is not found in memory or updating fails
func updateEventFromJsonData(gameVersion models.GameVersion, eventData EventFileData) error {
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

	// Atualizar o evento no datastore após modificações
	SetEvent(gameVersion, eventData.ID, eventFile)

	return nil
}

// updateEventStringFromJson updates a single string within an event based on JSON data.
// This function handles the string-level updates including localization processing.
//
// Parameters:
//   - eventFile: The event file object to update
//   - eventString: The string data from JSON
//   - eventID: The event ID (for logging purposes)
//
// Returns: error if string index is invalid or update fails
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

// updateEventStringLocalization updates a single localization for an event string object.
// This function handles the low-level localization update logic for event strings.
//
// Parameters:
//   - objToEdit: The string object to update
//   - localization: The localization code (e.g., "us", "jp")
//   - newString: The new string content
//
// Returns: error if localization is not supported or update fails
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
