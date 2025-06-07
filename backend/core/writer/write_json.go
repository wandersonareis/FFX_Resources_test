package writer

import (
	"encoding/json"
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
