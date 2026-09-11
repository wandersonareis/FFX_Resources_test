package exporters

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
	"sort"
)

// getLocalizationKeys returns all available localization keys
// This function returns the localization keys from the common package
func getLocalizationKeys() []string {
	var keys []string
	for key := range common.SupportedLanguages {
		keys = append(keys, key)
	}
	return keys
}

// getSortedLocalizationKeys returns localization keys sorted alphabetically
func getSortedLocalizationKeys() []string {
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)
	return localizationKeys
}

// buildEventStringData creates EventStringData from an event string with all localizations
//
// Parameters:
//   - index: Index of the string in the event
//   - str: String object that implements GetLocalizedString method
//   - localizationKeys: List of localization keys to process
//
// Returns: EventStringData with index and all localized text
func buildEventStringData(index int, str interface{ GetLocalizedString(string) string }, localizationKeys []string) EventStringData {
	stringData := EventStringData{
		Index: index,
		Text:  make(map[string]string),
	}

	for _, langKey := range localizationKeys {
		value := str.GetLocalizedString(langKey)
		stringData.Text[langKey] = value
	}

	return stringData
}

// processEventFromMemory processes an event from memory and creates EventFileData
//
// Parameters:
//   - eventID: ID of the event to process
//   - localizationKeys: List of localization keys to include
//
// Returns: EventFileData pointer or nil if no valid data found
func currentGameVersion() common.GameVersion {
	return interactions.CurrentGameVersion()
}

func processEventFromMemory(eventID string, localizationKeys []string) *EventFileData {
	eventFile := event.GetEvent(currentGameVersion(), eventID)
	if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
		return nil
	}

	eventData := EventFileData{
		ID:      eventFile.ID,
		Strings: make([]EventStringData, 0, len(eventFile.Strings)),
	}

	for i, str := range eventFile.Strings {
		stringData := buildEventStringData(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	if len(eventData.Strings) == 0 {
		common.LogVerbose("No strings found for event %s, skipping", eventID)
		return nil
	}

	return &eventData
}

// processEventFromFile processes an event from file and creates EventFileData
//
// Parameters:
//   - eventID: ID of the event to process
//   - localizationKeys: List of localization keys to include
//
// Returns: EventFileData pointer or nil if no valid data found
func processEventFromFile(eventID string, localizationKeys []string) *EventFileData {
	if common.IsVerboseMode() {
		common.LogVerbose("Exporting event file to JSON: %s", eventID)
	}

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	eventFileStrings, err := event.ReadLocalizedEventStrings(eventID, common.ToInt(version))
	if err != nil {
		common.LogVerbose("Error loading localized strings: %v", err)
		return nil
	}

	if len(eventFileStrings) == 0 {
		return nil
	}

	eventData := EventFileData{
		ID:      eventID,
		Strings: make([]EventStringData, 0, len(eventFileStrings)),
	}

	for i, str := range eventFileStrings {
		stringData := buildEventStringData(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	return &eventData
}

// writeEventJSONFile writes event data to a JSON file
//
// Parameters:
//   - events: Slice of EventFileData to write
//   - fileName: Name of the JSON file to create
//
// Returns: error if file creation fails
func writeEventJSONFile(events []EventFileData, fileName string) error {
	if len(events) == 0 {
		common.LogVerbose("No events with string data found to export to JSON")
		return nil
	}

	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	filePath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	export := make([]models.EventFileExport, 0, len(events))
	for _, e := range events {
		strings := make([]models.EventStringDataExport, 0, len(e.Strings))
		for _, s := range e.Strings {
			strings = append(strings, models.EventStringDataExport{Index: s.Index, Text: s.Text})
		}
		export = append(export, models.EventFileExport{
			Metadata: models.NewFileMetadata(models.NewFileInfoFromPath(models.EventBinaryPath(e.ID))),
			ID:       e.ID,
			Strings:  strings,
		})
	}

	if err := models.SaveDataFile(export, filePath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}

	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: %d", len(events))
	return nil
}

// ExportAllEventsToJSON exports all event data to a JSON file with all localizations.
//
// This function processes all event entries and exports them to
// "events_all_localizations.json" with all available localizations.
// Only exports events that have string data (skips empty events).
//
// Returns: error if export fails or data is not loaded
func ExportAllEventsToJSON() error {
	fileName := "events_all_localizations.json"
	localizationKeys := getSortedLocalizationKeys()
	eventIDs := event.GetAllEventIDs(currentGameVersion())

	var allEvents []EventFileData
	var count int

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			common.LogVerbose("Skipping event %s: no strings found", eventID)
			continue
		}

		allEvents = append(allEvents, *eventData)
		count++
	}

	common.LogVerbose("Total events processed: %d from %d", count, len(eventIDs))
	return writeEventJSONFile(allEvents, fileName)
}

// ExportEventsForLocalizationToJSON exports event data for a specific language to a JSON file.
//
// This function processes all event entries and exports them to a JSON file
// named "events_{languageCode}.json" with the specified localization only.
//
// Parameters:
//   - languageCode: Language code for localization (e.g., "us", "jp")
//
// Returns: error if export fails or data is not loaded
func ExportEventsForLocalizationToJSON(languageCode string) error {
	eventIDs := event.GetAllEventIDs(currentGameVersion())
	localizationKeys := []string{languageCode}

	var allEvents []EventFileData

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			continue
		}

		allEvents = append(allEvents, *eventData)
	}

	if len(allEvents) == 0 {
		return fmt.Errorf("no events with string data found for localization %s", languageCode)
	}

	fileName := fmt.Sprintf("events_%s.json", languageCode)
	return writeEventJSONFile(allEvents, fileName)
}

// ExportSingleEventToJSON exports a single event's data to a JSON file with all localizations.
//
// This function processes the event identified by eventId and exports it to
// "event_{eventId}_all_localizations.json" with all available localizations.
//
// Parameters:
//   - eventId: Unique identifier for the event to export
//
// Returns: error if export fails or data is not loaded
func ExportSingleEventToJSON(eventId string) error {
	localizationKeys := getSortedLocalizationKeys()

	eventData := processEventFromFile(eventId, localizationKeys)
	if eventData == nil {
		return fmt.Errorf("no data found for event %s", eventId)
	}

	allEvents := []EventFileData{*eventData}
	fileName := "event_" + eventId + "_all_localizations.json"

	return writeEventJSONFile(allEvents, fileName)
}
