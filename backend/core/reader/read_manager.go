package reader

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

/*
EVENT READING FUNCTIONS
=======================

This file contains functions for reading Final Fantasy X event files:

1. ReadAllEvents(skipBlitzballEvents, print) - Reads all event files from the originals directory
   - Scans the event directory structure (event/obj/XX/XXXX/)
   - Optionally skips blitzball events ("bl" folder)
   - Stores all events in the global EVENTS map

2. ReadEventFull(eventId, print) - Reads a complete event file with localized strings
   - Reads the main .ebp event file
   - Loads associated localized string files (.bin)
   - Optionally outputs debug information and processes scripts

3. ReadEventFile(eventId, path, print) - Reads a single event file from a specific path
   - Creates an EventFile instance from the binary data
   - Used internally by ReadEventFull

4. ReadLocalizedStringFiles(basePath) - Reads localized string files
   - Placeholder implementation for reading .bin localization files
   - Returns LocalizedFieldStringObject array

Usage:
  ReadAllEvents(common.SkipBlitzballEvents, true)  // Load all events
  event := ReadEventFull("0601", true)             // Load specific event
  loadedEvent := EVENTS["0601"]                    // Access from global storage
*/

func GetLocalizationRoot(localization string) string {
	return common.PathFfxRoot + "new_" + localization + "pc/"
}

func PrepareCharset(charset string) error {
	replacements := loadCharReplacements(charset)

	path := filepath.Join(common.PathOriginalsRoot, "ffx_encoding", "ffxsjistbl_"+charset+".bin")
	filePath, err := components.ResolveFile(path, false)
	if err != nil {
		return err
	}
	data, err := components.ReadFile(filePath)
	if err != nil {
		return err
	}

	str := string(data)	
	runes := []rune(str)
	if len(replacements) > 0 {
		runes = applyReplacements(runes, replacements)
	}
	
	byteToChar, charToByte := buildMappings(runes)
	components.SetCharMap(charset, byteToChar, charToByte)
	return nil
}

func loadCharReplacements(charset string) map[rune]rune {
	path := filepath.Join(common.PathOriginalsRoot, "ffx_encoding", "char_replacements.json")
	resolvedFile , err := components.ResolveFile(path, false)
	if err != nil {
		fmt.Printf("Error resolving char replacements file: %v\n", err)
		return nil
	}
	data, err := components.ReadFile(resolvedFile)
	if err != nil {
		return nil
	}

	var config map[string]map[string]string
	if err := json.Unmarshal(data, &config); err != nil {
		return nil
	}
	charsetReplacements, ok := config[charset]
	if !ok {
		return nil
	}

	replacements := make(map[rune]rune)
	for oldChar, newChar := range charsetReplacements {
		oldCharRune := []rune(oldChar)
		newCharRune := []rune(newChar)
		if len(oldCharRune) == 1 && len(newCharRune) == 1{
			replacements[oldCharRune[0]] = newCharRune[0]
		} else {
			fmt.Printf("Warning: Invalid replacement for charset %s: %s -> %s\n", charset, oldChar, newChar)
		}
	}
	return replacements
}

func applyReplacements(runes []rune, replacements map[rune]rune) []rune {
	if len(replacements) == 0 {
		return runes
	}

	newRunes := make([]rune, len(runes))
	copy(newRunes, runes)

	for i, r := range newRunes {
		if newChar, ok := replacements[r]; ok {
			newRunes[i] = newChar
		}
	}
	return newRunes
}

func buildMappings(runes []rune) (map[uint]rune, map[rune]uint) {
	byteToChar := make(map[uint]rune)
	charToByte := make(map[rune]uint)
	for i, r := range runes {
		idx := uint(i + 0x30) // Start at 0x30
		byteToChar[idx] = r
		if _, exists := charToByte[r]; !exists {
			charToByte[r] = idx
		}
	}
	return byteToChar, charToByte
}

func ByteToCharLookup(charset string, idx uint) (rune, bool) {
	m, ok := components.ByteToCharMaps[charset]
	if !ok {
		return 0, false
	}
	r, exists := m[idx]
	return r, exists
}

// CharToByteLookup returns the byte value (int) mapped to r in the given charset.
func CharToByteLookup(charset string, r rune) (uint, bool) {
	m, ok := components.CharToByteMaps[charset]
	if !ok {
		return 0, false
	}
	b, exists := m[r]
	return b, exists
}

func PrepareStringMacros(filename, localization string, printOutput bool) {
	// Read the entire file as bytes
	data := components.FileToBytes(filename, true)

	// Parse chunks and build the MacroDictionaryFile
	mdf := components.NewMacroDictionaryFile(data, localization)

	// Publish all extracted MacroStrings into the global MacroLookup
	mdf.PublishStrings()

	// Optionally print the MacroDictionaryFile (uses String() method)
	if printOutput {
		fmt.Println(mdf)
	}
}

func InitializeInternals() {
	// 1. Prepare all character maps
	for _, cs := range common.Charsets {
		PrepareCharset(cs)
	}

	// 2. Prepare string macros for each localization
	for loc := range common.Localizations {
		newVar := GetLocalizationRoot(loc)
		path := filepath.Join(newVar, "menu", "macrodic.dcp")
		
		PrepareStringMacros(path, loc, false)
	}
}

func ReadEventFileBasic(filename string) error {
	// Read the file using the components package
	data, err := components.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Process the data (this will depend on the structure of your event files and your application's needs)
	// For demonstration, let's just print the data
	fmt.Printf("Data from event file %s: %d bytes\n", filename, len(data))

	return nil
}

func LoadAllEvents(directory string) error {
	// Read all files in the events directory
	files, err := os.ReadDir(directory)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Sort files by name
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	// Iterate over files and read each event file
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".evnt") {
			filePath := filepath.Join(directory, file.Name())
			if err := ReadEventFileBasic(filePath); err != nil {
				return fmt.Errorf("failed to read event file %s: %w", file.Name(), err)
			}
		}
	}

	return nil
}

// ReadAllEvents reads all event files from the originals event directory
// skipBlitzballEvents: if true, skips the "bl" (blitzball) folder
// print: if true, prints debug information
func ReadAllEvents(print bool) error {
	eventsFolder, err := components.ResolveFile(common.PathOriginalsEvent, false)
	if err != nil {
		return fmt.Errorf("failed to resolve events directory: %w", err)
	}

	// Check if events directory exists
	if !common.IsPathExists(eventsFolder) {
		if print {
			fmt.Println("Cannot locate events at:", eventsFolder)
		}
		return fmt.Errorf("events directory not found: %s", eventsFolder)
	}

	// Read directory contents
	entries, err := os.ReadDir(eventsFolder)
	if err != nil {
		if print {
			fmt.Println("Cannot list events:", err)
		}
		return fmt.Errorf("failed to read events directory: %w", err)
	}

	var eventFiles []string

	// Process each subdirectory (like "06", "07", etc.)
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip blitzball events if requested
		if common.SkipBlitzballEvents && entry.Name() == "bl" {
			continue
		}

		// Read subdirectory contents
		subPath := filepath.Join(eventsFolder, entry.Name())
		subEntries, err := os.ReadDir(subPath)
		if err != nil {
			continue
		}

		// Process each event directory
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() || strings.HasPrefix(subEntry.Name(), ".") {
				continue
			}

			eventFiles = append(eventFiles, subEntry.Name())
		}
	}

	// Sort event files
	sort.Strings(eventFiles)

	// Process each event file
	for _, eventId := range eventFiles {
		eventFile, err := ReadEventFull(eventId, print)
		if err != nil {
			return fmt.Errorf("failed to read event file %s: %w", eventId, err)
		}
		components.EVENTS[eventId] = eventFile

	}
	return nil
}

// ReadEventFull reads a complete event file with localized strings
// eventId: the event identifier (e.g., "0601")
// print: if true, prints debug information and processes script
func ReadEventFull(eventId string, print bool) (*components.EventFile, error) {
	if len(eventId) < 2 {
		if print {
			fmt.Printf("Invalid event ID: %s\n", eventId)
		}
		return nil, fmt.Errorf("invalid event ID: %s", eventId)
	}

	// Extract the first two characters for the folder structure
	shortened := eventId[:2]

	// Build the path: event/obj/XX/XXXX/XXXX.ebp
	midPath := shortened + "/" + eventId + "/" + eventId
	eventPath := common.PathOriginalsEvent + midPath + ".ebp"
	originalsPath, err := components.ResolveFile(eventPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve event file path: %w", err)
	}

	// Read the main event file
	eventFile := ReadEventFile(eventId, originalsPath, print)
	if eventFile == nil {
		return nil, fmt.Errorf("failed to read event file: %s", originalsPath)
	}

	// Read localized string files
	localizedStrings := components.ReadLocalizedStringFiles("event/obj_ps3/" + midPath + ".bin")
	if localizedStrings != nil {
		eventFile.AddLocalizations(localizedStrings)
	}

	if print {
		textOutputPath := common.PathTextOutputRoot + "event/obj/" + shortened + "/" + eventId + ".txt"
		eventFileString := eventFile.String()

		// Ensure output directory exists
		if err := common.EnsurePathExists(textOutputPath); err != nil {
			if print {
				fmt.Printf("Failed to ensure output directory exists: %s\n", err)
			}
		}

		if print {
			fmt.Printf("Processed event %s, output would go to: %s\n", eventId, textOutputPath)
			fmt.Printf("Event content preview: %s\n", eventFileString)
		}
	}

	return eventFile, nil
}

// ReadEventFile reads a single event file from the given path
// eventId: the event identifier
// path: the file path to read
// print: if true, prints debug information
func ReadEventFile(eventId, path string, print bool) *components.EventFile {
	// Check if file exists
	if !common.IsPathExists(path) {
		if print {
			fmt.Printf("Event file not found: %s\n", path)
		}
		return nil
	}

	// Read file bytes
	data, err := components.ReadFile(path)
	if err != nil {
		if print {
			fmt.Printf("Error reading event file %s: %v\n", path, err)
		}
		return nil
	}

	// Create EventFile from bytes
	eventFile := components.NewEventFile(eventId, data)

	if print {
		fmt.Printf("Successfully read event file: %s (%d bytes)\n", eventId, len(data))
	}

	return eventFile
}
