package writer

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
	"sort"
)

func currentGameVersion() common.GameVersion {
	return interactions.CurrentGameVersion()
}

type EventStringData struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
}

type EventFileData struct {
	ID      string            `json:"id"`
	Strings []EventStringData `json:"strings"`
}

func prepareOutputDirectory() (string, error) {
	path := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(path); err != nil {
		return "", fmt.Errorf("error creating output directory: %v", err)
	}
	return path, nil
}

func getSortedEventIDs() []string {
	eventIDs := event.GetAllEventIDs(currentGameVersion())
	sort.Strings(eventIDs)
	return eventIDs
}

func getSortedLocalizationKeys() []string {
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)
	return localizationKeys
}

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

func processEventFromMemory(eventID string, localizationKeys []string) *EventFileData {
	eventFile := event.GetEvent(currentGameVersion(), eventID)
	if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
		return nil
	}

	if common.IsVerboseMode() {
		fmt.Printf("Processing event file: %s\n", eventFile.ID)
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
		if common.IsVerboseMode() {
			fmt.Printf("No strings found for event %s, skipping...\n", eventID)
		}
		return nil
	}

	return &eventData
}

func writeJSONFile(events []EventFileData, fileName, outputPath string) error {
	if len(events) == 0 {
		fmt.Println("No events with string data found to export to JSON")
		return nil
	}

	filePath := filepath.Join(outputPath, common.WithVersionSuffix(fileName))

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

	if common.IsVerboseMode() {
		fmt.Printf("Arquivo JSON de eventos exportado: %s\n", filePath)
		fmt.Printf("Total de eventos exportados: %d\n", len(events))
	}

	return nil
}

// ExportAllLocalizationsToJSON writes event files as JSON for all localizations
// Creates JSON files with event strings for each language in the edits/ directory
//
// JSON Format:
//   - Array of event objects, each containing ID and strings array
//   - Each string object has index and localized text for each language
//   - Only exports events that have string data (skips empty events)
//
// ExportAllLocalizationsToJSON processes all event entries and exports them to a JSON file,
// including all localized string data for each event.
//
// It performs the following steps:
//   - Ensures the output directory exists.
//   - Retrieves and sorts localization keys for consistent language ordering.
//   - Retrieves and sorts event IDs to guarantee deterministic output.
//   - Iterates over each event, skipping those without string data.
//   - Constructs an event data structure that includes an ordered list of strings, where each string is
//     mapped to its localized versions based on the available language keys.
//   - Marshals the assembled event data to JSON with proper formatting (indentation).
//   - Writes the JSON data to a file named "events_all_localizations.json" in the output directory.
//   - Optionally prints the file path and the total count of exported events if the 'print' parameter is true.
//
// This function logs errors encountered during directory creation, JSON marshaling, or file writing,
// and terminates early if any such error occurs.
func ExportAllLocalizationsToJSON() {
	outputPath, err := prepareOutputDirectory()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	localizationKeys := getSortedLocalizationKeys()
	eventIDs := getSortedEventIDs()

	var allEvents []EventFileData
	var count int

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			if common.IsVerboseMode() {
				fmt.Printf("Skipping event %s: no strings found\n", eventID)
			}
			continue
		}

		allEvents = append(allEvents, *eventData)
		count++
	}

	if common.IsVerboseMode() {
		fmt.Printf("Total events processed: %d from %d\n", count, len(eventIDs))
	}

	if err := writeJSONFile(allEvents, "events_all_localizations.json", outputPath); err != nil {
		fmt.Printf("%v\n", err)
	}
}
