package writer

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
	"sort"
)

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
	eventIDs := make([]string, 0, len(components.EVENTS))
	for eventID := range components.EVENTS {
		eventIDs = append(eventIDs, eventID)
	}
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
	eventFile := components.EVENTS[eventID]
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

func processEventFromFile(eventID string, localizationKeys []string) *EventFileData {
	if common.IsVerboseMode() {
		fmt.Printf("Exporting event file to JSON: %s\n", eventID)
	}

	eventFileStrings, err := components.ReadLocalizedEventStrings(eventID)
	if err != nil {
		fmt.Printf("Erro ao carregar strings localizadas: %v\n", err)
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

func writeJSONFile(events []EventFileData, fileName, outputPath string) error {
	if len(events) == 0 {
		fmt.Println("No events with string data found to export to JSON")
		return nil
	}

	filePath := filepath.Join(outputPath, fileName)

	if err := common.SaveAsJSON(events, filePath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %v", filePath, err)
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

// ExportAllEventsToJSONForLocalization generates a JSON file containing localized event data
// for the specified language.
//
// The function performs the following steps:
// 1. Constructs the output directory path using the game files root, mods folder, and specific subdirectories.
// 2. Ensures that the output directory exists; if not, it attempts to create it.
// 3. Collects and sorts the event IDs from the global event components to enforce consistent processing order.
// 4. Iterates through each event, building a corresponding localized data structure:
//   - Skips events with no strings or where no localized value is available.
//   - Creates structured event data with an ID and a slice of localized string entries,
//     where each string entry holds its index and a mapping of language code to the localized text.
//
// 5. If no valid events are processed, the function logs this and exits.
// 6. Marshals the structured event data into a prettified JSON format.
// 7. Writes the resulting JSON to a file, naming it in accordance with the language code (e.g., "events_us.json").
// 8. Optionally prints success messages and a summary of exported events if the 'print' flag is enabled.
//
// Parameters:
//
//	languageCode: A string representing the target language code for localization (e.g., "us").
//
// The function handles errors by printing them to the standard output. No value is returned.
func ExportAllEventsToJSONForLocalization(languageCode string) {
	outputPath, err := prepareOutputDirectory()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	eventIDs := getSortedEventIDs()
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
		fmt.Printf("No events with string data found for localization %s\n", languageCode)
		return
	}

	fileName := fmt.Sprintf("events_%s.json", languageCode)
	if err := writeJSONFile(allEvents, fileName, outputPath); err != nil {
		fmt.Printf("%v\n", err)
	}
}

// ExportEventLocalizationToJSON exports localized event string data to a JSON file.
//
// This function processes the event identified by eventId by reading its localized strings,
// organizing them by localization keys, and writing the result as a formatted JSON file into
// a predetermined directory structure. It ensures that the output directory exists, retrieves
// and sorts the localization keys for a consistent output, and handles error scenarios throughout
// the process.
//
// Parameters:
//
//	eventId - A unique identifier for the event whose strings are to be exported.
//	print   - A boolean flag indicating whether to print informational messages to the console,
//	          including export confirmation and the count of exported events.
//
// Process Overview:
//  1. Constructs the output directory path by joining game root and subdirectories.
//  2. Ensures that the designated output directory exists.
//  3. Retrieves and sorts the list of localization keys for consistent ordering.
//  4. Reads the localized strings for the event; if an error occurs during reading, the function exits.
//  5. Iterates over each event string, mapping available localized versions (or defaulting to empty strings).
//  6. Aggregates the processed event data into a collection.
//  7. Marshals the collection to a formatted JSON string and writes it to a file named based on eventId.
//  8. Optionally prints detailed messages if the print flag is set, including the file path and number
//     of exported events.
//
// Error Handling:
//   - If errors occur during directory creation, reading localized strings, marshalling to JSON,
//     or writing to the file system, the function logs the error and terminates early.
func ExportEventLocalizationToJSON(eventId string) {
	outputPath, err := prepareOutputDirectory()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	localizationKeys := getSortedLocalizationKeys()

	eventData := processEventFromFile(eventId, localizationKeys)
	if eventData == nil {
		return
	}

	allEvents := []EventFileData{*eventData}
	fileName := "event_" + eventId + "_all_localizations.json"

	if err := writeJSONFile(allEvents, fileName, outputPath); err != nil {
		fmt.Printf("%v\n", err)
	}
}
