package writer

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/reader"
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

// getLocalizationKeys returns all available localization keys
// This function returns the localization keys from the common package
func getLocalizationKeys() []string {
	var keys []string
	for key := range common.Localizations {
		keys = append(keys, key)
	}
	return keys
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

// EventStringData represents a single string with all its localizations
type EventStringData struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
}

// EventFileData represents an event file with all its strings
type EventFileData struct {
	ID      string            `json:"id"`
	Strings []EventStringData `json:"strings"`
}

// WriteEventFileForAllLocalizationsJSON writes event files as JSON for all localizations
// Creates JSON files with event strings for each language in the edits/events/ directory
//
// Parameters:
//   - print: If true, prints the exported file paths for debugging
//
// JSON Format:
//   - Array of event objects, each containing ID and strings array
//   - Each string object has index and localized text for each language
//   - Only exports events that have string data (skips empty events)
func WriteEventFileForAllLocalizationsJSON(print bool) {
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

	var allEvents []EventFileData

	// Iterate through all loaded events in sorted order
	for _, eventID := range eventIDs {
		eventFile := components.EVENTS[eventID]
		if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
			continue
		}

		fmt.Printf("Exporting event file to JSON: %s\n", eventFile.ID)

		// Create event data structure
		eventData := EventFileData{
			ID:      eventFile.ID,
			Strings: make([]EventStringData, 0, len(eventFile.Strings)),
		}

		// Process each string with its localizations
		for i, str := range eventFile.Strings {
			stringData := EventStringData{
				Index: i,
				Text:  make(map[string]string),
			}

			// Add localized text for each language
			for _, langKey := range localizationKeys {
				value := str.GetLocalizedString(langKey)
				if value != "" {
					stringData.Text[langKey] = value
				} else {
					stringData.Text[langKey] = ""
				}
			}

			eventData.Strings = append(eventData.Strings, stringData)
		}

		// Skip if no strings were added
		if len(eventData.Strings) == 0 {
			continue
		}

		allEvents = append(allEvents, eventData)
	}

	// Skip if no events were processed
	if len(allEvents) == 0 {
		fmt.Println("No events with string data found to export to JSON")
		return
	}

	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(allEvents, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling events to JSON: %v\n", err)
		return
	}

	// Write JSON file
	fileName := "events_all_localizations.json"
	filePath := filepath.Join(path, fileName)

	err = components.WriteStringToFile(filePath, string(jsonData))
	if err != nil {
		fmt.Printf("Error writing JSON file %s: %v\n", filePath, err)
		return
	}

	if print {
		fmt.Printf("Arquivo JSON de eventos exportado: %s\n", filePath)
		fmt.Printf("Total de eventos exportados: %d\n", len(allEvents))
	}
}

// WriteEventFileForLocalizationJSON writes event file JSON for a specific localization
// Similar to WriteEventFileForAllLocalizationsJSON but for a single language
func WriteEventFileForLocalizationJSON(localization string, print bool) {
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

	var allEvents []EventFileData

	// Iterate through all loaded events in sorted order
	for _, eventID := range eventIDs {
		fmt.Printf("Processing event file: %s\n", eventID)
		eventFile := components.EVENTS[eventID]
		if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
			continue
		}

		// Create event data structure for single localization
		eventData := EventFileData{
			ID:      eventFile.ID,
			Strings: make([]EventStringData, 0, len(eventFile.Strings)),
		}

		// Process each string for this localization
		for i, str := range eventFile.Strings {
			value := str.GetLocalizedString(localization)

			stringData := EventStringData{
				Index: i,
				Text:  map[string]string{localization: value},
			}

			eventData.Strings = append(eventData.Strings, stringData)
		}

		// Skip if no strings were added
		if len(eventData.Strings) == 0 {
			continue
		}

		allEvents = append(allEvents, eventData)
	}

	// Skip if no events were processed
	if len(allEvents) == 0 {
		fmt.Printf("No events with string data found for localization %s\n", localization)
		return
	}

	// Marshal to JSON with proper formatting
	jsonData, err := json.MarshalIndent(allEvents, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling events to JSON: %v\n", err)
		return
	}

	// Write JSON file
	fileName := fmt.Sprintf("events_%s.json", localization)
	filePath := filepath.Join(path, fileName)

	err = components.WriteStringToFile(filePath, string(jsonData))
	if err != nil {
		fmt.Printf("Error writing JSON file %s: %v\n", filePath, err)
		return
	}

	if print {
		fmt.Printf("Arquivo JSON de eventos exportado (%s): %s\n", localization, filePath)
		fmt.Printf("Total de eventos exportados: %d\n", len(allEvents))
	}
}

// ExampleWriteManagerUsage demonstrates how to use the write manager functions
func ExampleWriteManagerUsage() {
	fmt.Println("=== Write Manager Usage Example ===")

	// Ensure events are loaded first
	err := reader.ReadAllEvents(false) // false = don't skip blitzball
	if err != nil {
		fmt.Printf("Error loading events: %v\n", err)
		return
	}

	// Example 1: Write CSV files for all localizations
	fmt.Println("\n1. Writing CSV files for all localizations:")
	WriteEventFileForAllLocalizations(true)

	// Example 2: Write JSON files for all localizations
	fmt.Println("\n2. Writing JSON files for all localizations:")
	WriteEventFileForAllLocalizationsJSON(true)

	// Example 3: Write CSV files for a specific localization
	fmt.Println("\n3. Writing CSV files for Japanese localization:")
	WriteEventFileForLocalization("jp", true)

	// Example 4: Write JSON files for a specific localization
	fmt.Println("\n4. Writing JSON files for Japanese localization:")
	WriteEventFileForLocalizationJSON("jp", true)

	// Example 5: Write files for English localization (both formats)
	fmt.Println("\n5. Writing files for English localization:")
	WriteEventFileForLocalization("us", true)
	WriteEventFileForLocalizationJSON("us", true)

	fmt.Println("\n=== Write operations completed ===")
}

// WriteAllEventsExample demonstrates how to export events to CSV
// This function loads all events and then exports them to CSV format
func WriteAllEventsExample() {
	fmt.Println("=== Exemplo de Exportação de Eventos para CSV ===")

	// First, load all events
	fmt.Println("Carregando todos os eventos...")
	err := reader.ReadAllEvents(false) // false = include blitzball events
	if err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	// Count loaded events
	eventCount := len(components.EVENTS)
	fmt.Printf("Carregados %d eventos\n", eventCount)

	if eventCount == 0 {
		fmt.Println("Nenhum evento foi carregado. Verifique se os arquivos de evento existem.")
		return
	}

	// Export all events to CSV
	fmt.Println("Exportando eventos para CSV...")
	WriteEventFileForAllLocalizations(true)

	fmt.Println("=== Exportação concluída ===")
}

// WriteAllEventsJSONExample demonstrates how to export events to JSON
// This function loads all events and then exports them to JSON format
func WriteAllEventsJSONExample() {
	fmt.Println("=== Exemplo de Exportação de Eventos para JSON ===")

	// First, load all events
	fmt.Println("Carregando todos os eventos...")
	err := reader.ReadAllEvents(false) // false = include blitzball events
	if err != nil {
		fmt.Printf("Erro ao carregar eventos: %v\n", err)
		return
	}

	// Count loaded events
	eventCount := len(components.EVENTS)
	fmt.Printf("Carregados %d eventos\n", eventCount)

	if eventCount == 0 {
		fmt.Println("Nenhum evento foi carregado. Verifique se os arquivos de evento existem.")
		return
	}

	// Export all events to JSON
	fmt.Println("Exportando eventos para JSON (todas as localizações)...")
	WriteEventFileForAllLocalizationsJSON(true)

	// Export specific localizations to JSON
	fmt.Println("Exportando eventos para JSON (localização específica - JP)...")
	WriteEventFileForLocalizationJSON("jp", true)

	fmt.Println("Exportando eventos para JSON (localização específica - US)...")
	WriteEventFileForLocalizationJSON("us", true)

	fmt.Println("=== Exportação JSON concluída ===")
}
