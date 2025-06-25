package reader

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func EditAndSaveEventCSVFiles() error {
	csvPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")

	if !common.IsPathExists(csvPath) {
		common.LogVerbose("Directory not found: %s", csvPath)
		return fmt.Errorf("directory not found: %s", csvPath)
	}

	err := filepath.Walk(csvPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".csv") {
			common.LogVerbose("Processing CSV file: %s", path)
			return editAndSaveEventFromCSV(path)
		}

		return nil
	})

	if err != nil {
		common.LogVerbose("Error processing CSV files: %v", err)
		return err
	}

	return nil
}

func editAndSaveEventFromCSV(csvPath string) error {
	lines, err := csvToList(csvPath)
	if err != nil {
		common.LogVerbose("Error reading CSV file %s: %v", csvPath, err)
		return err
	}

	if len(lines) <= 1 {
		common.LogVerbose("CSV file is empty or has only header: %s", csvPath)
		return nil
	}

	header := lines[0]
	idCol := -1
	stringIndexCol := -1
	colToLocalization := make(map[int]string)

	for i, col := range header {
		colLower := strings.ToLower(strings.TrimSpace(col))
		switch colLower {
		case "string index":
			stringIndexCol = i
		case "id":
			idCol = i
		default:
			if _, exists := common.SupportedLanguages[colLower]; exists {
				colToLocalization[i] = colLower
			}
		}
	}

	if idCol < 0 || stringIndexCol < 0 {
		common.LogVerbose("Required columns not found in file: %s (id=%d, string index=%d)", csvPath, idCol, stringIndexCol)
		return nil
	}

	values := lines[1:] // Skip header
	processedEventIDs := make(map[string]bool)

	for _, cells := range values {
		if len(cells) <= idCol || len(cells) <= stringIndexCol {
			continue
		}

		eventID := strings.TrimSpace(cells[idCol])
		stringIndexStr := strings.TrimSpace(cells[stringIndexCol])

		stringIndex, err := strconv.Atoi(stringIndexStr)
		if err != nil {
			common.LogVerbose("Invalid string index '%s' in file %s", stringIndexStr, csvPath)
			continue
		}

		eventFile, exists := components.EVENTS[eventID]
		if !exists || eventFile == nil {
			common.LogVerbose("Event not found: %s", eventID)
			continue
		}

		if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
			common.LogVerbose("String index out of range for event %s: %d", eventID, stringIndex)
			continue
		}

		processedEventIDs[eventID] = true

		objToEdit := eventFile.Strings[stringIndex]
		if common.IsVerboseMode() {
			var localizedStrings []string
			for colIdx := range colToLocalization {
				if colIdx < len(cells) {
					localizedStrings = append(localizedStrings, cells[colIdx])
				}
			}
			fmt.Printf("Copying [\"%s\"] into %s[%d]\n",
				strings.Join(localizedStrings, "\",\""), eventID, stringIndex)

		}

		for colIdx, localization := range colToLocalization {
			if colIdx < len(cells) {
				newString := strings.TrimSpace(cells[colIdx])
				if newString != "" {
					fieldString := objToEdit.GetLocalizedContent(localization)
					if fieldString != nil {
						fieldString.SetRegularString(newString)
					}
				}
			}
		}
	}

	for eventID := range processedEventIDs {
		if err := ExportEventStringsToLocalizations(eventID); err != nil {
			common.LogVerbose("Error saving event %s: %v", eventID, err)
			return fmt.Errorf("error saving event %s: %w", eventID, err)
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

func ExportEventStringsToLocalizations(eventID string) error {
	eventFile, exists := components.EVENTS[eventID]
	if !exists || eventFile == nil {
		return fmt.Errorf("event not found: %s", eventID)
	}

	if len(eventID) < 2 {
		return fmt.Errorf("invalid event ID: %s", eventID)
	}

	pathPattern := filepath.Join("event/obj_ps3/", eventID[:2], eventID, eventID+".bin")

	return writeStringsToStringToFileForAllLocalizations(pathPattern, eventFile.Strings)
}

func writeStringsToStringToFileForAllLocalizations(pathPattern string, localizedStrings []*components.LocalizedFieldStringObject) error {
	common.LogVerbose("Writing string file: %s", pathPattern)

	for localizationKey := range common.SupportedLanguages {
		localizationRoot := common.GetLocalizationRoot(localizationKey)

		localePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, localizationRoot, pathPattern)
		localePath = filepath.FromSlash(localePath)

		stringsBytes := stringsToStringFileBytes(localizedStrings, localizationKey)

		dir := filepath.Dir(localePath)
		if err := common.EnsurePathExists(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		if err := common.WriteBytesToFile(localePath, stringsBytes); err != nil {
			return fmt.Errorf("failed to write file %s: %w", localePath, err)
		}
		common.LogVerbose("Wrote localized strings to %s (%d bytes)", localePath, len(stringsBytes))
	}
	return nil
}

func stringsToStringFileBytes(localizedStrings []*components.LocalizedFieldStringObject, languageCode string) []byte {
	if len(localizedStrings) == 0 {
		return []byte{}
	}

	charset := components.GetCharsetForLanguage(languageCode)
	fieldStrings := make([]*components.FieldString, 0, len(localizedStrings))

	for _, localizedObj := range localizedStrings {
		if localizedObj == nil {
			continue
		}
		fieldString := localizedObj.GetLocalizedContent(languageCode)
		if fieldString == nil {
			fieldString = &components.FieldString{Charset: charset}
		}
		fieldStrings = append(fieldStrings, fieldString)
	}

	stringBytes := components.RebuildFieldStrings(fieldStrings, charset)

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

func EditAndSaveEventJSONFiles() error {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")

	if !common.IsPathExists(jsonPath) {
		return fmt.Errorf("directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, "events_all_localizations.json")

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Processing JSON file: %s", jsonFilePath)

	if err := editAndSaveEventFromJSON(jsonFilePath); err != nil {
		common.LogVerbose("Error processing JSON file: %v", err)
		return err
	}
	common.LogVerbose("Events processed successfully!")
	return nil
}

func editAndSaveEventFromJSON(jsonPath string) error {
	allEvents, err := common.ReadJsonFile[[]EventFileDataJSON](jsonPath)
	if err != nil {
		fmt.Printf("Erro ao ler arquivo JSON %s: %v\n", jsonPath, err)
		return err
	}

	processedEventIDs := make(map[string]bool)

	for _, eventData := range allEvents {
		eventFile, exists := components.EVENTS[eventData.ID]
		if !exists || eventFile == nil {
			common.LogVerbose("Event not found: %s", eventData.ID)
			continue
		}

		processedEventIDs[eventData.ID] = true
		common.LogVerbose("Processing event %s with %d strings", eventData.ID, len(eventData.Strings))

		for _, eventString := range eventData.Strings {
			stringIndex := eventString.Index
			common.LogVerbose("Processing string %d for event %s", stringIndex, eventData.ID)

			if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
				common.LogVerbose("String index out of range for event %s: %d", eventData.ID, stringIndex)
				continue
			}

			objToEdit := eventFile.Strings[stringIndex]

			common.LogVerbose("Updating event %s[%d] with %d localizations",
				eventData.ID, stringIndex, len(eventString.Text))

			for localization, newString := range eventString.Text {
				if newString != "" {
					if _, exists := common.SupportedLanguages[localization]; exists {
						fieldString := objToEdit.GetLocalizedContent(localization)
						if fieldString != nil {
							fieldString.SetRegularString(newString)
						}
					}
				}
			}
		}
	}

	for eventID := range processedEventIDs {
		if err := ExportEventStringsToLocalizations(eventID); err != nil {
			return fmt.Errorf("error saving event %s: %w", eventID, err)
		}
	}

	return nil
}

// EditAndSaveSpecificEventFromJSON processes a specific event from the events_all_localizations.json file
// This function loads the JSON file, finds the specified event, and applies changes only to that event
//
// Parameters:
//   - eventID: The ID of the event to process (e.g., "ev001", "btl_001")
//
// Returns:
//   - error: nil if successful, error if the event is not found or processing fails
func EditAndSaveSpecificEventFromJSON(eventID string) error {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	jsonFilePath := filepath.Join(jsonPath, "events_all_localizations.json")

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Loading JSON file: %s", jsonFilePath)
	common.LogVerbose("Looking for event: %s", eventID)

	file, err := os.Open(jsonFilePath)
	if err != nil {
		common.LogVerbose("Error opening JSON file %s: %v", jsonFilePath, err)
		return err
	}
	defer file.Close()

	// Parse JSON content - expecting array of EventFileDataJSON
	var allJsonEvents []EventFileDataJSON
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&allJsonEvents); err != nil {
		common.LogVerbose("Error decoding JSON %s: %v", jsonFilePath, err)
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
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	common.LogVerbose("Event %s found in JSON with %d strings", eventID, len(targetEventData.Strings))

	// Validate event exists in memory
	eventFile, exists := components.EVENTS[eventID]
	if !exists || eventFile == nil {
		return fmt.Errorf("event not found in memory: %s", eventID)
	}

	common.LogVerbose("Processing event %s with %d strings", eventID, len(targetEventData.Strings))

	// Process each event string
	for _, eventString := range targetEventData.Strings {
		stringIndex := eventString.Index

		common.LogVerbose("Processing string %d for event %s", stringIndex, eventID)

		if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
			common.LogVerbose("String index out of range for event %s: %d", eventID, stringIndex)
			continue
		}

		objToEdit := eventFile.Strings[stringIndex]

		common.LogVerbose("Updating event %s[%d] with %d localizations",
			eventID, stringIndex, len(eventString.Text))

		for localization, newString := range eventString.Text {
			if newString != "" {
				if _, exists := common.SupportedLanguages[localization]; exists {
					fieldString := objToEdit.GetLocalizedContent(localization)
					if fieldString != nil {
						fieldString.SetRegularString(newString)
					}
				}
			}
		}
	}

	// Write updated event back to files
	if err := ExportEventStringsToLocalizations(eventID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventID, err)
		return err
	}

	common.LogVerbose("Event %s processed and saved successfully", eventID)
	return nil
}

// EventFileDataJSON represents the JSON structure for a single event file
// This matches the format exported by WriteEventFileForAllLocalizationsJSON
type EventFileDataJSON = EventFileData

// EventStringDataJSON represents a single event string with its localizations
// This matches the format exported by WriteEventFileForAllLocalizationsJSON
//type EventStringDataJSON = EventStringData

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

// MacroStringData represents a single macro string with its text variations
type MacroStringData struct {
	Index          int    `json:"index"`
	RegularText    string `json:"regular_text"`
	SimplifiedText string `json:"simplified_text"`
	HasDistinct    bool   `json:"has_distinct_simplified"`
}

// MacroChunkData represents a chunk of macro strings
type MacroChunkData struct {
	ChunkIndex int               `json:"chunk_index"`
	Strings    []MacroStringData `json:"strings"`
}

// MacroLocalizationData represents macro dictionary data for a specific localization
type MacroLocalizationData struct {
	Localization string           `json:"localization"`
	Chunks       []MacroChunkData `json:"chunks"`
}

/*
JSON MACRO DICTIONARY EDITOR FUNCTIONS
======================================

This section contains JSON equivalents for macro dictionary editor functions.
These functions process the single JSON file created by WriteMacroDictionaryJSON.

1. EditAndSaveMacroDictJSONFiles(print) - Processes the macro_dictionary_all_localizations.json file
   - Reads the specific JSON file created by WriteMacroDictionaryJSON
   - Processes all macro dictionaries from the single JSON file
   - Applies changes back to the MACRODICTFILE component
   - Uses internal implementation to avoid circular dependencies

2. editAndSaveMacroDictFromJSON(print, path) - Processes a single JSON file
   - Reads JSON content and parses it internally
   - Updates MACRODICTFILE with macro dictionary data
   - Reconstructs MacroString objects from JSON data

3. EditAndSaveSpecificMacroDictFromJSON(localization, print) - Processes a specific localization
   - Loads the macro_dictionary_all_localizations.json file
   - Searches for the specified localization by code
   - Processes only that localization and applies changes back to MACRODICTFILE

Usage:
  EditAndSaveMacroDictJSONFiles(true)  // Process macro_dictionary_all_localizations.json with debug output
  EditAndSaveMacroDictJSONFiles(false) // Process silently

  EditAndSaveSpecificMacroDictFromJSON("us", true)  // Process only US localization with debug output
  EditAndSaveSpecificMacroDictFromJSON("jp", false) // Process only Japanese localization silently
*/
func EditAndSaveMacroDictJSONFiles() error {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", "macrodic")

	if !common.IsPathExists(jsonPath) {
		return fmt.Errorf("directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, "macro_dictionary_all_localizations.json")

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Processing macro dictionary JSON file: %s", jsonFilePath)

	err := editAndSaveMacroDictFromJSON(jsonFilePath)
	if err != nil {
		common.LogVerbose("Error processing macro dictionary JSON file: %v", err)
		return err
	}

	common.LogVerbose("Macro dictionary processed successfully")
	return nil
}

func editAndSaveMacroDictFromJSON(jsonPath string) error {
	common.LogVerbose("Loading macro dictionary data from JSON file: %s", jsonPath)

	// Read JSON file
	resolvedFile, err := common.NewFileAccessor(jsonPath)
	if err != nil {
		return fmt.Errorf("error resolving JSON file path: %v", err)
	}
	/* jsonData, err := common.ReadFile(resolvedFile.ResolvedPath)
	if err != nil {
		return fmt.Errorf("error reading JSON file: %v", err)
	} */

	// Try to parse as array of localizations first (all localizations file)
	allLocalizations, err := common.ReadJsonFile[[]MacroLocalizationData](resolvedFile.ResolvedPath)
	/* var allLocalizations []MacroLocalizationData
	if err := json.Unmarshal(jsonData, &allLocalizations); err != nil {
		// If that fails, try to parse as single localization
		var singleLocalization MacroLocalizationData
		if err := json.Unmarshal(jsonData, &singleLocalization); err != nil {
			return fmt.Errorf("error parsing JSON: %v", err)
		}
		// Convert single localization to array
		allLocalizations = []MacroLocalizationData{singleLocalization}
	} */

	common.LogVerbose("Found %d localizations in JSON file", len(allLocalizations))
	var debugEntries []DebugMacroDicEntry

	// Process each localization
	for _, locData := range allLocalizations {
		/* if locData.Localization != "us" {
			continue
		} */
		common.LogVerbose("Processing localization: %s (%d chunks)", locData.Localization, len(locData.Chunks))

		// Clear existing data for this localization
		//components.MACRODICTFILE[locData.Localization] = make([][]*components.MacroString, 0)
		macroCharsetStrings := components.MACRODICTFILE[locData.Localization]

		// Find the maximum chunk index to properly size the array
		maxChunkIndex := 15

		// Initialize the chunks array with proper size
		chunks := make([][]*components.MacroString, maxChunkIndex+1)
		for i := range chunks {
			chunks[i] = make([]*components.MacroString, 0)
		}

		// Process each chunk
		for _, chunkData := range locData.Chunks {
			chunkIndex := chunkData.ChunkIndex

			common.LogVerbose("Processing chunk %d (%d strings)", chunkIndex, len(chunkData.Strings))

			// Find the maximum string index to properly size the chunk
			maxStringIndex := -1
			for _, stringData := range chunkData.Strings {
				if stringData.Index > maxStringIndex {
					maxStringIndex = stringData.Index
				}
			}

			// Initialize the strings array for this chunk
			if maxStringIndex >= 0 {
				chunks[chunkIndex] = make([]*components.MacroString, maxStringIndex+1)
			}

			charset := components.GetCharsetForLanguage(locData.Localization)
			for _, stringData := range chunkData.Strings {
				stringIndex := stringData.Index

				simplifiedText := stringData.SimplifiedText
				var regularBytes []byte
				var simplifiedBytes []byte

				regularBytes = components.StringToBytes(stringData.RegularText, charset)
				bytesToString := components.BytesToString(regularBytes, charset)

				// Verificar se bytesToString é diferente de stringData.RegularText
				if bytesToString != stringData.RegularText {
					// Obter dados originais do macroCharsetStrings
					var originalBytes []byte
					var originalString string

					if chunkIndex < len(macroCharsetStrings) && stringIndex < len(macroCharsetStrings[chunkIndex]) {
						originalMacroString := macroCharsetStrings[chunkIndex][stringIndex]
						if originalMacroString != nil {
							originalBytes = originalMacroString.GetRegularBytes()
							originalString = originalMacroString.GetRegularString()
						}
					} // Criar entrada de debug
					debugEntry := DebugMacroDicEntry{
						Localization:      locData.Localization,
						ChunkIndex:        chunkIndex,
						StringIndex:       stringIndex,
						OriginalBytes:     hex.EncodeToString(originalBytes),
						OriginalString:    originalString,
						TextJSON:          stringData.RegularText,
						ConvertedTextJSON: bytesToString,
					}
					debugEntries = append(debugEntries, debugEntry)
				}

				if stringData.HasDistinct {
					simplifiedBytes = components.StringToBytes(simplifiedText, charset)
				}

				// Create MacroString object
				macroString := &components.MacroString{
					Charset:          charset,
					RegularOffset:    0, // These offsets are not relevant when reconstructing from JSON
					SimplifiedOffset: 0, // They are used during binary parsing only
					RegularBytes:     regularBytes,
					SimplifiedBytes:  simplifiedBytes,
				}

				if stringIndex >= len(chunks[chunkIndex]) {
					newSize := stringIndex + 1
					newSlice := make([]*components.MacroString, newSize)
					copy(newSlice, chunks[chunkIndex])
					chunks[chunkIndex] = newSlice
				}
				chunks[chunkIndex][stringIndex] = macroString
			}
			components.RebuildMacroStrings(chunks[chunkIndex], charset, false)
		}

		// Update the global MACRODICTFILE
		components.MACRODICTFILE[locData.Localization] = chunks
		common.LogVerbose("Localization %s updated successfully with %d chunks", locData.Localization, len(chunks))
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Macro dictionary data loaded successfully")

		totalLocalizations := len(components.MACRODICTFILE)
		totalStrings := 0

		for localization, chunks := range components.MACRODICTFILE {
			localizationStrings := 0
			for _, chunk := range chunks {
				for _, macroString := range chunk {
					if macroString != nil && !macroString.IsEmpty() {
						localizationStrings++
					}
				}
			}
			totalStrings += localizationStrings
			common.LogVerbose("- Localization %s: %d chunks, %d strings", localization, len(chunks), localizationStrings)
		}
		common.LogVerbose("Total: %d localizations, %d macro strings loaded", totalLocalizations, totalStrings)
	}

	// Criar arquivo de debug se houver entradas
	if len(debugEntries) > 0 {
		if err := createDebugMacroDicFile(debugEntries); err != nil {
			common.LogVerbose("Warning: could not create debug file: %v", err)
		}
	}

	return nil
}

// EditAndSaveSpecificMacroDictFromJSON processes a specific localization from the macro_dictionary_all_localizations.json file
// This function loads the JSON file, finds the specified localization, and applies changes only to that localization
//
// Parameters:
//   - localization: The localization code to process (e.g., "us", "jp", "de", etc.)
//
// Returns:
//   - error: Any error that occurred during processing
func EditAndSaveSpecificMacroDictFromJSON(localization string) error {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	jsonFilePath := filepath.Join(jsonPath, "macro_dictionary_all_localizations.json")

	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("arquivo JSON não encontrado: %s", jsonFilePath)
	}

	common.LogVerbose("Processing specific localization %s from macro dictionary JSON file: %s",
		localization, jsonFilePath)

	// Read JSON file
	resolvedFile, err := common.NewFileAccessor(jsonFilePath)
	if err != nil {
		return fmt.Errorf("error resolving JSON file path: %v", err)
	}
	allLocalizations, err := common.ReadJsonFile[[]MacroLocalizationData](resolvedFile.ResolvedPath)
	if err != nil {
		return fmt.Errorf("error on reading JSON file: %v", err)
	}

	/* jsonData, err := common.ReadFile(resolvedFile.ResolvedPath)
	if err != nil {
		return fmt.Errorf("erro ao ler arquivo JSON: %v", err)
	}

	// Parse JSON to find all localizations
	var allLocalizations []MacroLocalizationData
	if err := json.Unmarshal(jsonData, &allLocalizations); err != nil {
		// If that fails, try to parse as single localization
		var singleLocalization MacroLocalizationData
		if err := json.Unmarshal(jsonData, &singleLocalization); err != nil {
			return fmt.Errorf("erro ao fazer parse do JSON: %v", err)
		}
		// Convert single localization to array
		allLocalizations = []MacroLocalizationData{singleLocalization}
	} */

	// Find the specific localization
	var targetLocalization *MacroLocalizationData
	for i, locData := range allLocalizations {
		if locData.Localization == localization {
			targetLocalization = &allLocalizations[i]
			break
		}
	}

	if targetLocalization == nil {
		return fmt.Errorf("location %s not found in the json file", localization)
	}

	common.LogVerbose("Location %s found with %d chunks", localization, len(targetLocalization.Chunks))

	// Clear existing data for this localization only
	components.MACRODICTFILE[localization] = make([][]*components.MacroString, 0)

	// Find the maximum chunk index to properly size the array
	maxChunkIndex := 15

	// Initialize the chunks array with proper size
	chunks := make([][]*components.MacroString, maxChunkIndex+1)
	for i := range chunks {
		chunks[i] = make([]*components.MacroString, 0)
	}

	// Process each chunk for the target localization
	for _, chunkData := range targetLocalization.Chunks {
		chunkIndex := chunkData.ChunkIndex

		common.LogVerbose("Processing chunk %d (%d strings)", chunkIndex, len(chunkData.Strings))

		// Find the maximum string index to properly size the chunk
		maxStringIndex := -1
		for _, stringData := range chunkData.Strings {
			if stringData.Index > maxStringIndex {
				maxStringIndex = stringData.Index
			}
		}

		// Initialize the strings array for this chunk
		for _, stringData := range chunkData.Strings {
			stringIndex := stringData.Index

			// Get the text for this localization
			regularText := stringData.RegularText
			simplifiedText := stringData.SimplifiedText

			// If no simplified text is provided, use regular text
			if simplifiedText == "" {
				simplifiedText = regularText
			}

			// Convert strings back to bytes using the localization's charset
			charset := components.GetCharsetForLanguage(localization)
			regularBytes := components.StringToBytes(regularText, charset)
			simplifiedBytes := components.StringToBytes(simplifiedText, charset)

			// Create MacroString object
			macroString := &components.MacroString{
				Charset:          charset,
				RegularOffset:    0, // These offsets are not relevant when reconstructing from JSON
				SimplifiedOffset: 0, // They are used during binary parsing only
				RegularBytes:     regularBytes,
				SimplifiedBytes:  simplifiedBytes,
			}

			// Add to chunk (ensure the array is large enough)
			if stringIndex >= len(chunks[chunkIndex]) {
				// Expand the array to accommodate this index
				newSize := stringIndex + 1
				newSlice := make([]*components.MacroString, newSize)
				copy(newSlice, chunks[chunkIndex])
				chunks[chunkIndex] = newSlice
			}
			chunks[chunkIndex][stringIndex] = macroString
		}
	}

	// Update the global MACRODICTFILE for this specific localization only
	components.MACRODICTFILE[localization] = chunks

	if common.IsVerboseMode() {
		stringCount := 0
		for _, chunk := range chunks {
			for _, macroString := range chunk {
				if macroString != nil && !macroString.IsEmpty() {
					stringCount++
				}
			}
		}
		common.LogVerbose("✓ Location %S successfully processed: %of chunks, %d strings",
			localization, len(chunks), stringCount)
	}

	return nil
}

// Estrutura para armazenar dados de debug de comparação de strings
type DebugMacroDicEntry struct {
	Localization      string `json:"localization"`
	ChunkIndex        int    `json:"chunkIndex"`
	StringIndex       int    `json:"stringIndex"`
	OriginalBytes     string `json:"originalBytes"` // Hex representation
	OriginalString    string `json:"originalString"`
	TextJSON          string `json:"textJson"`
	ConvertedTextJSON string `json:"convertedTextJson"`
}

type DebugMacroDicEntries struct {
	ErrorsCount int                  `json:"errorsCount"`
	Entries     []DebugMacroDicEntry `json:"entries"`
}

// Função para criar arquivo de debug com dados de comparação de strings
func createDebugMacroDicFile(entries []DebugMacroDicEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Criar nome do arquivo com data e hora
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("debug_macrodic_%s.json", timestamp)
	debugEntries := DebugMacroDicEntries{
		ErrorsCount: len(entries),
		Entries:     entries,
	}
	// Criar o arquivo JSON
	jsonData, err := json.MarshalIndent(debugEntries, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar dados de debug: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de debug %s: %v", filename, err)
	}

	fmt.Printf("Arquivo de debug criado: %s com %d entradas\n", filename, len(entries))
	return nil
}
