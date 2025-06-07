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
	resolvedFile, err := components.ResolveFile(path, false)
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
		if len(oldCharRune) == 1 && len(newCharRune) == 1 {
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

func CharToByteLookup(charset string, r rune) (uint, bool) {
	m, ok := components.CharToByteMaps[charset]
	if !ok {
		return 0, false
	}
	b, exists := m[r]
	return b, exists
}

func PrepareStringMacros(filename, localization string) {
	data := components.FileToBytes(filename, true)

	mdf := components.NewMacroDictionaryFile(data, localization)

	mdf.PublishStrings()
}

func InitializeInternals() error {
	for _, cs := range common.Charsets {
		if err := PrepareCharset(cs); err != nil {
			return err
		}
	}

	for loc := range common.Localizations {
		newVar := GetLocalizationRoot(loc)
		path := filepath.Join(newVar, "menu", "macrodic.dcp")
		
		PrepareStringMacros(path, loc)
	}
	for idx, strings := range components.MacroLookup {
		fmt.Printf("MacroLookup[%d] s%dl%d\n", idx, idx/0x100, idx%0x100)
		for loc := range common.Localizations {
			content := strings.GetLocalizedContent(loc)
			if content == nil || content.IsEmpty() {
				continue
			}
			macro := content.GetString()
			fmt.Printf("  %s: %s\n", loc, macro)
			}
		}
	return nil
}

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

		subPath := filepath.Join(eventsFolder, entry.Name())
		subEntries, err := os.ReadDir(subPath)
		if err != nil {
			continue
		}

		for _, subEntry := range subEntries {
			if !subEntry.IsDir() || strings.HasPrefix(subEntry.Name(), ".") {
				continue
			}

			eventFiles = append(eventFiles, subEntry.Name())
		}
	}

	sort.Strings(eventFiles)

	for _, eventId := range eventFiles {
		eventFile, err := ReadEventFull(eventId, print)
		if err != nil {
			return fmt.Errorf("failed to read event file %s: %w", eventId, err)
		}
		components.EVENTS[eventId] = eventFile

	}
	return nil
}

func ReadEventFull(eventId string, print bool) (*components.EventFile, error) {
	if len(eventId) < 2 {
		if print {
			fmt.Printf("Invalid event ID: %s\n", eventId)
		}
		return nil, fmt.Errorf("invalid event ID: %s", eventId)
	}

	shortened := eventId[:2]

	midPath := shortened + "/" + eventId + "/" + eventId
	eventPath := common.PathOriginalsEvent + midPath + ".ebp"
	originalsPath, err := components.ResolveFile(eventPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve event file path: %w", err)
	}

	eventFile := ReadEventFile(eventId, originalsPath, print)
	if eventFile == nil {
		return nil, fmt.Errorf("failed to read event file: %s", originalsPath)
	}

	localizedStrings := components.ReadLocalizedStringFiles("event/obj_ps3/" + midPath + ".bin")
	if localizedStrings != nil {
		eventFile.AddLocalizations(localizedStrings)
	}

	if print {
		textOutputPath := common.PathTextOutputRoot + "event/obj/" + shortened + "/" + eventId + ".txt"
		eventFileString := eventFile.String()

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
