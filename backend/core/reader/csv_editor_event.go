package reader

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

/*
CSV & JSON EVENT EDITOR FUNCTIONS
=================================

This file contains functions for reading and processing CSV and JSON event files:

CSV FUNCTIONS:
1. EditAndSaveEventCSVFiles(print) - Processes all CSV files in the events directory
   - Scans the edits/events/ directory for .csv files
   - Processes each CSV file to update event strings
   - Applies changes back to the original event files

2. editAndSaveEventFromCSV(print, path) - Processes a single CSV file
   - Reads CSV content and parses header to identify columns
   - Maps CSV data back to EventFile strings
   - Updates localized content for each language
   - Saves changes to event files

JSON FUNCTIONS:
3. EditAndSaveEventJSONFiles(print) - Processes the events_all_localizations.json file
   - Reads the specific JSON file created by WriteEventFileForAllLocalizationsJSON
   - Processes all events from the single JSON file
   - Applies changes back to the original event files

4. EditAndSaveSpecificEventFromJSON(eventID, print) - Processes a specific event from JSON
   - Loads the events_all_localizations.json file
   - Searches for the specified event by ID
   - Processes only that event and applies changes back to files

5. editAndSaveEventFromJSON(print, path) - Processes a single JSON file
   - Reads JSON content and parses structure
   - Maps JSON data back to EventFile strings
   - Updates localized content for each language
   - Saves changes to event files

Usage:
  CSV Workflow:
    EditAndSaveEventCSVFiles(true)   // Process all CSV files with debug output
    EditAndSaveEventCSVFiles(false)  // Process silently

  JSON Workflow:
    EditAndSaveEventJSONFiles(true)  // Process events_all_localizations.json with debug output
    EditAndSaveEventJSONFiles(false) // Process silently

  Specific Event JSON Workflow:
    EditAndSaveSpecificEventFromJSON("ev001", true)  // Process specific event with debug output
    EditAndSaveSpecificEventFromJSON("btl_001", false) // Process specific event silently
*/

func EditAndSaveEventCSVFiles(print bool) error {
	csvPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "events")

	if !common.IsPathExists(csvPath) {
		fmt.Printf("Diretório não encontrado: %s\n", csvPath)
		return fmt.Errorf("directory not found: %s", csvPath)
	}

	err := filepath.Walk(csvPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".csv") {
			if print {
				fmt.Printf("Processando arquivo: %s\n", path)
			}
			return editAndSaveEventFromCSV(print, path)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Erro ao processar arquivos CSV: %v\n", err)
		return err
	}

	return nil
}

// editAndSaveEventFromCSV processes a single CSV file and applies changes to the corresponding event
func editAndSaveEventFromCSV(print bool, csvPath string) error {
	// Read CSV file
	lines, err := csvToList(csvPath)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo CSV %s: %v\n", csvPath, err)
		return err
	}

	// Need at least header + 1 data row
	if len(lines) <= 1 {
		if print {
			fmt.Printf("Arquivo CSV vazio ou só com cabeçalho: %s\n", csvPath)
		}
		return nil
	}

	// Parse header to identify columns
	header := lines[0]
	idCol := -1
	stringIndexCol := -1
	colToLocalization := make(map[int]string)

	for i, col := range header {
		colLower := strings.ToLower(strings.TrimSpace(col))

		// Handle "string index" column
		if colLower == "string index" {
			stringIndexCol = i
		} else if colLower == "id" {
			idCol = i
		} else {
			// Check if it's a localization key
			if _, exists := common.Localizations[colLower]; exists {
				colToLocalization[i] = colLower
			}
		}
	}

	// Validate required columns
	if idCol < 0 || stringIndexCol < 0 {
		if print {
			fmt.Printf("Colunas obrigatórias não encontradas no arquivo: %s (id=%d, string index=%d)\n",
				csvPath, idCol, stringIndexCol)
		}
		return nil
	}

	// Process data rows
	values := lines[1:] // Skip header
	processedEventIDs := make(map[string]bool)

	for _, cells := range values {
		// Validate row has enough columns
		if len(cells) <= idCol || len(cells) <= stringIndexCol {
			continue
		}

		eventID := strings.TrimSpace(cells[idCol])
		stringIndexStr := strings.TrimSpace(cells[stringIndexCol])

		// Parse string index
		stringIndex, err := strconv.Atoi(stringIndexStr)
		if err != nil {
			if print {
				fmt.Printf("Índice de string inválido '%s' no arquivo %s\n", stringIndexStr, csvPath)
			}
			continue
		}

		// Get event from global EVENTS map
		eventFile, exists := components.EVENTS[eventID]
		if !exists || eventFile == nil {
			if print {
				fmt.Printf("Evento não encontrado: %s\n", eventID)
			}
			continue
		}

		// Validate string index
		if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
			if print {
				fmt.Printf("Índice de string fora do range para evento %s: %d\n", eventID, stringIndex)
			}
			continue
		}

		// Mark this event as processed
		processedEventIDs[eventID] = true

		// Get the string object to edit
		objToEdit := eventFile.Strings[stringIndex]
		// Build debug string if needed
		if print {
			var localizedStrings []string
			for colIdx := range colToLocalization {
				if colIdx < len(cells) {
					localizedStrings = append(localizedStrings, cells[colIdx])
				}
			}
			fmt.Printf("Copying [\"%s\"] into %s[%d]\n",
				strings.Join(localizedStrings, "\",\""), eventID, stringIndex)
		}

		// Update localized content
		for colIdx, localization := range colToLocalization {
			if colIdx < len(cells) {
				newString := strings.TrimSpace(cells[colIdx])
				if newString != "" {
					// Get the localized content for this language
					fieldString := objToEdit.GetLocalizedContent(localization)
					if fieldString != nil {
						fieldString.SetRegularString(newString)
					}
				}
			}
		}
	}

	// Write updated events back to files
	for eventID := range processedEventIDs {
		err := WriteEventStringsForAllLocalizations(eventID, print)
		if err != nil {
			fmt.Printf("Erro ao salvar evento %s: %v\n", eventID, err)
		}
	}

	return nil
}

func csvToList(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func WriteEventStringsForAllLocalizations(eventID string, print bool) error {
	eventFile, exists := components.EVENTS[eventID]
	if !exists || eventFile == nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	if len(eventID) < 2 {
		return fmt.Errorf("invalid event ID: %s", eventID)
	}

	pathPattern := "event/obj_ps3/" + eventID[:2] + "/" + eventID + "/" + eventID + ".bin"

	return writeStringFileForAllLocalizations(pathPattern, eventFile.Strings, print)
}

func writeStringFileForAllLocalizations(pathPattern string, localizedStrings []*components.LocalizedFieldStringObject, print bool) error {
	if print {
		fmt.Printf("Writing string file: %s\n", pathPattern)
	}

	for localizationKey := range common.Localizations {
		localizationRoot := GetLocalizationRoot(localizationKey)
		localePath := filepath.Join(components.GameFilesRoot, components.ModsFolder, localizationRoot, pathPattern)

		localePath = filepath.FromSlash(localePath)

		stringsBytes := stringsToStringFileBytes(localizedStrings, localizationKey)

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		err := components.WriteByteArrayToFile(localePath, stringsBytes)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", localePath, err)
		}

		if print {
			fmt.Printf("  Written: %s\n", localePath)
		}
	}

	return nil
}

func stringsToStringFileBytes(localizedStrings []*components.LocalizedFieldStringObject, localizationKey string) []byte {
	if len(localizedStrings) == 0 {
		return []byte{}
	}

	fieldStrings := make([]*components.FieldString, 0, len(localizedStrings))
	charset := components.LocalizationToCharset(localizationKey)

	for _, localizedObj := range localizedStrings {
		if localizedObj != nil {
			fieldString := localizedObj.GetLocalizedContent(localizationKey)
			if fieldString != nil {
				fieldStrings = append(fieldStrings, fieldString)
			} else {
				// Create empty field string if no content for this localization
				emptyFieldString := &components.FieldString{
					Charset: charset,
				}
				fieldStrings = append(fieldStrings, emptyFieldString)
			}
		}
	}

	stringBytes := components.RebuildFieldStrings(fieldStrings, charset, true)

	var buf bytes.Buffer

	for _, str := range fieldStrings {
		if str != nil {
			buf.Write(str.ToRegularHeaderBytes())

			buf.Write(str.ToSimplifiedHeaderBytes())
		}
	}

	buf.Write(stringBytes)

	return buf.Bytes()
}

/*
JSON EVENT EDITOR FUNCTIONS
===========================

This section contains JSON equivalents for CSV event editor functions.
These functions process the single JSON file created by WriteEventFileForAllLocalizationsJSON.

1. EditAndSaveEventJSONFiles(print) - Processes the events_all_localizations.json file
   - Reads the specific JSON file created by WriteEventFileForAllLocalizationsJSON
   - Processes all events from the single JSON file
   - Applies changes back to the original event files

2. editAndSaveEventFromJSON(print, path) - Processes a single JSON file
   - Reads JSON content and parses it into EventFileData structure
   - Maps JSON data back to EventFile strings
   - Updates localized content for each language
   - Saves changes to event files

Usage:
  EditAndSaveEventJSONFiles(true)  // Process events_all_localizations.json with debug output
  EditAndSaveEventJSONFiles(false) // Process silently
*/

func EditAndSaveEventJSONFiles(print bool) error {
	jsonPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "events")

	if !common.IsPathExists(jsonPath) {
		fmt.Printf("Diretório não encontrado: %s\n", jsonPath)
		return fmt.Errorf("directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, "events_all_localizations.json")

	if !common.IsPathExists(jsonFilePath) {
		fmt.Printf("Arquivo JSON não encontrado: %s\n", jsonFilePath)
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	if print {
		fmt.Printf("Processando arquivo JSON: %s\n", jsonFilePath)
	}

	err := editAndSaveEventFromJSON(print, jsonFilePath)
	if err != nil {
		fmt.Printf("Erro ao processar arquivo JSON: %v\n", err)
		return err
	}

	return nil
}

func editAndSaveEventFromJSON(print bool, jsonPath string) error {
	file, err := os.Open(jsonPath)
	if err != nil {
		fmt.Printf("Erro ao abrir arquivo JSON %s: %v\n", jsonPath, err)
		return err
	}
	defer file.Close()

	var allEvents []EventFileDataJSON
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&allEvents); err != nil {
		fmt.Printf("Erro ao decodificar JSON %s: %v\n", jsonPath, err)
		return err
	}

	processedEventIDs := make(map[string]bool)

	for _, eventData := range allEvents {
		eventFile, exists := components.EVENTS[eventData.ID]
		if !exists || eventFile == nil {
			if print {
				fmt.Printf("Evento não encontrado: %s\n", eventData.ID)
			}
			continue
		}

		processedEventIDs[eventData.ID] = true
		fmt.Printf("Processando evento %s com %d strings\n", eventData.ID, len(eventData.Strings))

		for _, eventString := range eventData.Strings {
			stringIndex := eventString.Index
			fmt.Printf("Processando string %d para evento %s\n", stringIndex, eventData.ID)

			if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
				if print {
					fmt.Printf("Índice de string fora do range para evento %s: %d\n", eventData.ID, stringIndex)
				}
				continue
			}

			objToEdit := eventFile.Strings[stringIndex]

			if print {
				fmt.Printf("Atualizando evento %s[%d] com %d localizações\n",
					eventData.ID, stringIndex, len(eventString.Text))
			}

			// Update localized content
			for localization, newString := range eventString.Text {
				if newString != "" {
					// Verify localization exists
					if _, exists := common.Localizations[localization]; exists {
						// Get the localized content for this language
						fieldString := objToEdit.GetLocalizedContent(localization)
						if fieldString != nil {
							fieldString.SetRegularString(newString)
						}
					}
				}
			}
		}
	}

	// Write updated events back to files
	for eventID := range processedEventIDs {
		err := WriteEventStringsForAllLocalizations(eventID, print)
		if err != nil {
			fmt.Printf("Erro ao salvar evento %s: %v\n", eventID, err)
		}
	}

	return nil
}

// EditAndSaveSpecificEventFromJSON processes a specific event from the events_all_localizations.json file
// This function loads the JSON file, finds the specified event, and applies changes only to that event
//
// Parameters:
//   - eventID: The ID of the event to process (e.g., "ev001", "btl_001")
//   - print: If true, prints debug information during processing
//
// Returns:
//   - error: nil if successful, error if the event is not found or processing fails
func EditAndSaveSpecificEventFromJSON(eventID string, print bool) error {
	jsonPath := filepath.Join(components.GameFilesRoot, components.ModsFolder, "edits", "events")
	jsonFilePath := filepath.Join(jsonPath, "events_all_localizations.json")

	// Check if the specific JSON file exists
	if !common.IsPathExists(jsonFilePath) {
		fmt.Printf("Arquivo JSON não encontrado: %s\n", jsonFilePath)
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	if print {
		fmt.Printf("Carregando arquivo JSON: %s\n", jsonFilePath)
		fmt.Printf("Procurando evento: %s\n", eventID)
	}

	file, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Printf("Erro ao abrir arquivo JSON %s: %v\n", jsonFilePath, err)
		return err
	}
	defer file.Close()

	// Parse JSON content - expecting array of EventFileDataJSON
	var allJsonEvents []EventFileDataJSON
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&allJsonEvents); err != nil {
		fmt.Printf("Erro ao decodificar JSON %s: %v\n", jsonFilePath, err)
		return err
	}

	// Find the specific event in the JSON
	var targetEventData *EventFileDataJSON
	for i := range allJsonEvents {
		if allJsonEvents[i].ID == eventID {
			targetEventData = &allJsonEvents[i]
			break
		}
	}

	// Check if event was found in JSON
	if targetEventData == nil {
		fmt.Printf("Evento %s não encontrado no arquivo JSON\n", eventID)
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	if print {
		fmt.Printf("Evento %s encontrado no JSON com %d strings\n", eventID, len(targetEventData.Strings))
	}

	// Validate event exists in memory
	eventFile, exists := components.EVENTS[eventID]
	if !exists || eventFile == nil {
		fmt.Printf("Evento não encontrado na memória: %s\n", eventID)
		return fmt.Errorf("event not found in memory: %s", eventID)
	}

	if print {
		fmt.Printf("Processando evento %s com %d strings\n", eventID, len(targetEventData.Strings))
	}

	// Process each event string
	for _, eventString := range targetEventData.Strings {
		stringIndex := eventString.Index

		if print {
			fmt.Printf("Processando string %d para evento %s\n", stringIndex, eventID)
		}

		// Validate string index
		if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
			if print {
				fmt.Printf("Índice de string fora do range para evento %s: %d\n", eventID, stringIndex)
			}
			continue
		}

		// Get the string object to edit
		objToEdit := eventFile.Strings[stringIndex]

		if print {
			fmt.Printf("Atualizando evento %s[%d] com %d localizações\n",
				eventID, stringIndex, len(eventString.Text))
		}

		// Update localized content
		for localization, newString := range eventString.Text {
			if newString != "" {
				// Verify localization exists
				if _, exists := common.Localizations[localization]; exists {
					// Get the localized content for this language
					fieldString := objToEdit.GetLocalizedContent(localization)
					if fieldString != nil {
						fieldString.SetRegularString(newString)
						if print {
							fmt.Printf("  Atualizado %s: %s\n", localization, newString)
						}
					}
				} else if print {
					fmt.Printf("  Localização não reconhecida: %s\n", localization)
				}
			}
		}
	}

	// Write updated event back to files
	err = WriteEventStringsForAllLocalizations(eventID, print)
	if err != nil {
		fmt.Printf("Erro ao salvar evento %s: %v\n", eventID, err)
		return err
	}

	if print {
		fmt.Printf("Evento %s processado e salvo com sucesso\n", eventID)
	}

	return nil
}

// EventFileDataJSON represents the JSON structure for a single event file
// This matches the format exported by WriteEventFileForAllLocalizationsJSON
type EventFileDataJSON = EventFileData

// EventStringDataJSON represents a single event string with its localizations
// This matches the format exported by WriteEventFileForAllLocalizationsJSON
type EventStringDataJSON = EventStringData

// EventFileData represents an event file with all its strings (same as writer package)
type EventFileData struct {
	ID      string            `json:"id"`
	Strings []EventStringData `json:"strings"`
}

// EventStringData represents a single event string with its localizations (same as writer package)
type EventStringData struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
}

// ExampleCsvEditorUsage demonstrates how to use the CSV editor functions
// EditAndSaveEventCsv is a wrapper for EditAndSaveEventCSVFiles for backward compatibility
// Deprecated: Use EditAndSaveEventCSVFiles instead
func EditAndSaveEventCsv(print bool) error {
	return EditAndSaveEventCSVFiles(print)
}
