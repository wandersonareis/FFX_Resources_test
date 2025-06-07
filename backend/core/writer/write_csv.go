package writer

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// WriteEventFileForAllLocalizations writes event files as CSV for all localizations
// Creates CSV files with event strings for each language in the edits/events/ directory
//
// Parameters:
//   - print: If true, prints the exported file paths for debugging
//
// CSV Format:
//   - Columns: id, string index, [language codes...]
//   - Each row represents one string with its translations across all languages
//   - Only exports events that have string data (skips empty events)
func WriteEventFileForAllLocalizations(print bool) {
	path := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "events")

	// Ensure the output directory exists
	if err := common.EnsurePathExists(path); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}
	// Get all localization keys and sort them for consistent output
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)

	// Get sorted list of event IDs for consistent ordering
	eventIDs := make([]string, 0, len(components.EVENTS))
	for eventID := range components.EVENTS {
		eventIDs = append(eventIDs, eventID)
	}
	sort.Strings(eventIDs)

	// Iterate through all loaded events in sorted order
	for _, eventID := range eventIDs {
		eventFile := components.EVENTS[eventID]
		if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
			continue
		}

		fmt.Printf("Exporting event file: %s\n", eventFile.ID)

		var csvBuilder strings.Builder

		// Write CSV header
		csvBuilder.WriteString("id,string index")
		for _, langKey := range localizationKeys {
			csvBuilder.WriteString(",")
			csvBuilder.WriteString(langKey)
		}
		csvBuilder.WriteString("\n")

		// Write each string with its localizations
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

		// Skip if CSV content is too small (just headers)
		if csvBuilder.Len() <= 40 {
			continue
		}

		// Write CSV file
		fileName := eventFile.ID + ".csv"
		filePath := filepath.Join(path, fileName)

		err := components.WriteStringToFile(filePath, csvBuilder.String())
		if err != nil {
			fmt.Printf("Error writing CSV file %s: %v\n", filePath, err)
			continue
		}

		if print {
			fmt.Printf("Arquivo CSV de eventos exportado: %s\n", filePath)
		}
	}
}

// WriteEventFileForLocalization writes event file CSV for a specific localization
// Similar to WriteEventFileForAllLocalizations but for a single language
func WriteEventFileForLocalization(localization string, print bool) {
	path := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "events")
	// Ensure the output directory exists
	if err := common.EnsurePathExists(path); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Get sorted list of event IDs for consistent ordering
	eventIDs := make([]string, 0, len(components.EVENTS))
	for eventID := range components.EVENTS {
		eventIDs = append(eventIDs, eventID)
	}
	sort.Strings(eventIDs)

	for _, eventID := range eventIDs {
		eventFile := components.EVENTS[eventID]
		if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
			continue
		}

		var csvBuilder strings.Builder

		// Write CSV header for single localization
		csvBuilder.WriteString("id,string index,")
		csvBuilder.WriteString(localization)
		csvBuilder.WriteString("\n")

		// Write each string for this localization
		for i, str := range eventFile.Strings {
			csvBuilder.WriteString(escapeCsvValue(eventFile.ID))
			csvBuilder.WriteString(",")
			csvBuilder.WriteString(fmt.Sprintf("%d", i))
			csvBuilder.WriteString(",")

			value := str.GetLocalizedString(localization)
			if value != "" {
				csvBuilder.WriteString(escapeCsvValue(value))
			}
			csvBuilder.WriteString("\n")
		}

		// Skip if CSV content is too small
		if csvBuilder.Len() <= 40 {
			continue
		}

		// Write CSV file
		fileName := eventFile.ID + "_" + localization + ".csv"
		filePath := filepath.Join(path, fileName)

		err := components.WriteStringToFile(filePath, csvBuilder.String())
		if err != nil {
			fmt.Printf("Error writing CSV file %s: %v\n", filePath, err)
			continue
		}

		if print {
			fmt.Printf("Arquivo CSV de eventos exportado (%s): %s\n", localization, filePath)
		}
	}
}

// escapeCsvValue escapes a string value for CSV format
// Handles quotes, commas, and newlines according to CSV standards
func escapeCsvValue(value string) string {
	// If value contains comma, quote, or newline, wrap in quotes and escape internal quotes
	if strings.Contains(value, ",") || strings.Contains(value, "\"") || strings.Contains(value, "\n") || strings.Contains(value, "\r") {
		// Escape internal quotes by doubling them
		escaped := strings.ReplaceAll(value, "\"", "\"\"")
		return "\"" + escaped + "\""
	}
	return value
}
