package writer

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func prepareOutputDirectoryCSV() (string, error) {
	path := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(path); err != nil {
		return "", fmt.Errorf("error creating output directory: %v", err)
	}
	return path, nil
}

func getSortedEventIDsCSV() []string {
	eventIDs := make([]string, 0, len(components.EVENTS))
	for eventID := range components.EVENTS {
		eventIDs = append(eventIDs, eventID)
	}
	sort.Strings(eventIDs)
	return eventIDs
}

func getSortedLocalizationKeysCSV() []string {
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)
	return localizationKeys
}

func buildCSVHeader(localizationKeys []string) string {
	var header strings.Builder
	header.WriteString("id,string index")
	for _, langKey := range localizationKeys {
		header.WriteString(",")
		header.WriteString(langKey)
	}
	header.WriteString("\n")
	return header.String()
}

func buildCSVRowForEvent(eventFile *components.EventFile, localizationKeys []string) string {
	var csvBuilder strings.Builder

	for i, str := range eventFile.Strings {
		csvBuilder.WriteString(escapeCsvValue(eventFile.ID))
		csvBuilder.WriteString(",")
		csvBuilder.WriteString(fmt.Sprintf("%d", i))

		for _, langKey := range localizationKeys {
			value := str.GetLocalizedString(langKey)
			csvBuilder.WriteString(",")
			if value != "" {
				csvBuilder.WriteString(escapeCsvValue(value))
			} else {
				csvBuilder.WriteString(escapeCsvValue(""))
			}
		}
		csvBuilder.WriteString("\n")
	}

	return csvBuilder.String()
}

func processEventFromMemoryCSV(eventID string, localizationKeys []string) string {
	eventFile := components.EVENTS[eventID]
	if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
		return ""
	}

	if common.IsVerboseMode() {
		fmt.Printf("Processing event file: %s\n", eventFile.ID)
	}

	return buildCSVRowForEvent(eventFile, localizationKeys)
}

func processEventFromFileCSV(eventID string, localizationKeys []string) string {
	if common.IsVerboseMode() {
		fmt.Printf("Exporting event file to CSV: %s\n", eventID)
	}

	eventFileStrings, err := components.ReadLocalizedEventStrings(eventID)
	if err != nil {
		fmt.Printf("Erro ao carregar strings localizadas: %v\n", err)
		return ""
	}

	if len(eventFileStrings) == 0 {
		return ""
	}

	// Create a temporary event file structure
	eventFile := &components.EventFile{
		ID:      eventID,
		Strings: eventFileStrings,
	}

	return buildCSVRowForEvent(eventFile, localizationKeys)
}

func writeCSVFile(csvContent, fileName, outputPath string, eventsWritten int) error {
	if eventsWritten == 0 {
		fmt.Printf("No events with string data found to export\n")
		return nil
	}

	filePath := filepath.Join(outputPath, fileName)

	err := common.WriteStringToFile(filePath, csvContent)
	if err != nil {
		return fmt.Errorf("error writing consolidated CSV file %s: %v", filePath, err)
	}

	fmt.Printf("Successfully exported %d events to consolidated CSV: %s\n", eventsWritten, filePath)
	if common.IsVerboseMode() {
		fmt.Printf("Arquivo CSV consolidado de eventos exportado: %s\n", filePath)
	}

	return nil
}

// ExportAllEventsToCSV writes all event files into a single CSV for all localizations
// Creates one consolidated CSV file with event strings for all languages in the edits/ directory
//
// CSV Format:
//   - Columns: id, string index, [language codes...]
//   - Each row represents one string with its translations across all languages
//   - All events are combined into a single file for easier management
//   - Only includes events that have string data (skips empty events)
func ExportAllEventsToCSV() {
	outputPath, err := prepareOutputDirectoryCSV()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	localizationKeys := getSortedLocalizationKeysCSV()
	eventIDs := getSortedEventIDsCSV()

	fmt.Printf("Exporting all events to single CSV file...\n")

	var csvBuilder strings.Builder
	csvBuilder.WriteString(buildCSVHeader(localizationKeys))

	eventsWritten := 0

	for _, eventID := range eventIDs {
		csvRow := processEventFromMemoryCSV(eventID, localizationKeys)
		if csvRow == "" {
			continue
		}

		csvBuilder.WriteString(csvRow)
		eventsWritten++
	}

	fileName := "events_all_localizations.csv"
	if err := writeCSVFile(csvBuilder.String(), fileName, outputPath, eventsWritten); err != nil {
		fmt.Printf("%v\n", err)
	}
}

// ExportAllEventsToCsvForLocalization writes all event files into a single CSV for a specific localization
// Creates one consolidated CSV file with event strings for the specified language in the edits/ directory
//
// CSV Format:
//   - Columns: id, string index, [localization]
//   - Each row represents one string with its translation for the specified language
//   - All events are combined into a single file for easier management
//   - Only includes events that have string data (skips empty events)
//
// Parameters:
//   - localization: The localization code (e.g., "en", "pt", "es") to export strings for
//
// The function handles errors gracefully and provides feedback on the export process.
// In verbose mode, it prints the path of the successfully exported CSV file.
func ExportAllEventsToCsvForLocalization(localization string) {
	outputPath, err := prepareOutputDirectoryCSV()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	eventIDs := getSortedEventIDsCSV()
	localizationKeys := []string{localization}

	fmt.Printf("Exporting all events to single CSV file for localization: %s\n", localization)

	var csvBuilder strings.Builder
	csvBuilder.WriteString(buildCSVHeader(localizationKeys))

	eventsWritten := 0

	for _, eventID := range eventIDs {
		csvRow := processEventFromMemoryCSV(eventID, localizationKeys)
		if csvRow == "" {
			continue
		}

		csvBuilder.WriteString(csvRow)
		eventsWritten++
	}

	if eventsWritten == 0 {
		fmt.Printf("No events with string data found to export for localization: %s\n", localization)
		return
	}

	fileName := "events_all_" + localization + "_localization.csv"
	if err := writeCSVFile(csvBuilder.String(), fileName, outputPath, eventsWritten); err != nil {
		fmt.Printf("%v\n", err)
	}
}

// ExportEventLocalizationToCSV exports localized event string data to a CSV file.
//
// This function processes the event identified by eventId by reading its localized strings,
// organizing them by localization keys, and writing the result as a formatted CSV file into
// a predetermined directory structure. It ensures that the output directory exists, retrieves
// and sorts the localization keys for a consistent output, and handles error scenarios throughout
// the process.
//
// Parameters:
//
//	eventId - A unique identifier for the event whose strings are to be exported.
//
// Process Overview:
//  1. Constructs the output directory path by joining game root and subdirectories.
//  2. Ensures that the designated output directory exists.
//  3. Retrieves and sorts the list of localization keys for consistent ordering.
//  4. Reads the localized strings for the event; if an error occurs during reading, the function exits.
//  5. Iterates over each event string, mapping available localized versions.
//  6. Writes the CSV content to a file named based on eventId.
//  7. Optionally prints detailed messages if verbose mode is enabled.
//
// Error Handling:
//   - If errors occur during directory creation, reading localized strings, or writing to the file system,
//     the function logs the error and terminates early.
func ExportEventLocalizationToCSV(eventId string) {
	outputPath, err := prepareOutputDirectoryCSV()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	localizationKeys := getSortedLocalizationKeysCSV()

	var csvBuilder strings.Builder
	csvBuilder.WriteString(buildCSVHeader(localizationKeys))

	csvRow := processEventFromFileCSV(eventId, localizationKeys)
	if csvRow == "" {
		fmt.Printf("No string data found for event %s\n", eventId)
		return
	}

	csvBuilder.WriteString(csvRow)

	fileName := "event_" + eventId + "_all_localizations.csv"
	if err := writeCSVFile(csvBuilder.String(), fileName, outputPath, 1); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func escapeCsvValue(value string) string {
	if strings.Contains(value, ",") || strings.Contains(value, "\"") || strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		escaped := strings.ReplaceAll(value, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return value
}
