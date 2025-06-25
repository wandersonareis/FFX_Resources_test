package converter

import (
	"encoding/binary"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadNameOnlyDataObjectsWithIlist reads binary data from a specified pattern path
// and creates NameOnlyDataObject entries with all available localizations.
//
// This function reads LocalizedTextObject entries containing only name information
// for various game elements that don't require descriptions. Each entry includes
// localized text for all supported languages in the game, making it suitable for
// simple text elements like labels, button text, or single-word entries.
//
// File format: name only data (binary format)
// Pattern path: Variable, passed as parameter (e.g., "battle/kernel/btl_txt.bin")
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: IList[ILocalizedTextObject] containing NameOnlyDataObject entries with full localization data
func ReadNameOnlyDataObjectsWithIlist(patternPath string) components.IList[components.ILocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
		return components.NewNameOnlyDataObject(data, stringBytes, headerLength, localization)
	}

	nameObjects := ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)

	populateDataObjectLocalizationsWithIlist(patternPath, nameObjects, creator)

	common.LogVerbose("Loading %d name-only data objects...\n", nameObjects.GetLength())

	return nameObjects
}

// ReadNameDescriptionObjectsWithIlist reads binary data from a specified pattern path
// and creates NameDescriptionTextObject entries with all available localizations.
//
// This function reads LocalizedTextObject entries containing both name and description information
// for various game elements that require detailed text. Each entry includes localized text for all
// supported languages in the game, making it suitable for complex game objects like items, commands,
// abilities, and other elements that need both a title and detailed description.
//
// File format: name and description data (binary format)
// Pattern path: Variable, passed as parameter (e.g., "battle/kernel/command.bin")
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: IList[ILocalizedTextObject] containing NameDescriptionTextObject entries with full localization data
func ReadNameDescriptionObjectsWithIlist(patternPath string) components.IList[components.ILocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
		return components.NewNameDescriptionTextObject(data, stringBytes, headerLength, localization)
	}

	nameDescObjects := ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)

	populateDataObjectLocalizationsWithIlist(patternPath, nameDescObjects, creator)

	common.LogVerbose("Loading %d name and description objects...\n", nameDescObjects.GetLength())

	return nameDescObjects
}

// ReadNameOnlyObjects reads binary data from a specified pattern path and creates
// NameOnlyTextObject entries for the default localization only.
//
// This function is used for legacy compatibility and creates slice-based results
// instead of IList-based results. For new code, prefer ReadNameOnlyDataObjectsWithIlist.
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: Slice of NameOnlyTextObject pointers with default localization only
/* func ReadNameOnlyObjects(patternPath string) components.IList[components.ILocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) components.ILocalizedTextObject {
		return components.NewNameOnlyTextObject(data, stringBytes, headerLength, localization)
	}

	nameObjects := ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)

	common.LogVerbose("Loading %d name-only objects...\n", nameObjects.GetLength())

	return nameObjects
} */

// ReadNameDescriptionObjects reads binary data from a specified pattern path and creates
// NameDescriptionTextObject entries for the default localization only.
//
// This function is used for legacy compatibility and creates slice-based results
// instead of IList-based results. For new code, prefer ReadNameDescriptionObjectsWithIlist.
//
// Parameters:
//   - patternPath: Relative path to the binary file within the localization directory
//
// Returns: Slice of NameDescriptionTextObject pointers with default localization only
/* func ReadNameDescriptionObjects(patternPath string) []*components.NameDescriptionTextObject {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, languageCode string) *components.NameDescriptionTextObject {
		return components.NewNameDescriptionTextObject(data, stringBytes, headerLength, languageCode)
	}

	commonObjects := components.ReadDataArray(filePath, common.DefaultLocalization, creator)

	common.LogVerbose("Loading %d name and description objects...\n", commonObjects.GetLength())

	return commonObjects.GetItems()
} */

// populateDataObjectLocalizationsWithIlist populates localization data for all supported languages
// by reading corresponding binary files for each language and merging the localized content
// into the existing LocalizedTextObject instances within the provided IList.
//
// This function iterates through all supported languages, reads the binary files for each
// language, and applies the localized text content to the corresponding objects in the IList,
// leveraging IList's GetItems() method for efficient access to the underlying collection.
//
// Parameters:
//   - path: Relative path to the binary file pattern (e.g., "battle/kernel/command.bin")
//   - objects: IList[ILocalizedTextObject] containing instances to be populated with localizations
//   - creator: Function that creates LocalizedTextObject instances from binary data
func populateDataObjectLocalizationsWithIlist(path string, objects components.IList[components.ILocalizedTextObject], creator func([]byte, []byte, int, string) components.ILocalizedTextObject) {
	if objects.IsEmpty() {
		return
	}

	for locKey := range common.SupportedLanguages {
		fullPath := filepath.Join(common.GetLocalizationRoot(locKey), path)

		localizations := ReadDataListWithIlist(fullPath, locKey, creator)
		if localizations != nil {
			maxLen := min(localizations.GetLength(), objects.GetLength())
			items := objects.GetItems()
			locItems := localizations.GetItems()
			for i := range maxLen {
				items[i].SetLocalizations(locItems[i])
			}
		}
	}
}

// ReadDataListWithIlist reads binary data files and creates localized text objects using a custom creator function.
//
// This is a low-level function that provides flexible binary data reading for any type of localized text object.
// It handles the standard binary format used by the game's localization files, which includes a header section
// with object metadata and data sections containing object-specific information and string data.
//
// The function reads the binary file structure:
// - 8 bytes: header information
// - 2 bytes: minimum index value
// - 2 bytes: maximum index value
// - 2 bytes: individual object length
// - 2 bytes: total data section length
// - 4 bytes: reserved/padding
// - Variable: object data section
// - Variable: string data section
//
// Parameters:
//   - filename: Path to the binary file to read
//   - languageCode: Language code for localization (e.g., "us", "jp", "de")
//   - creator: Function that creates ILocalizedTextObject instances from binary data
//     Parameters: (objData []byte, stringBytes []byte, headerLength int, languageCode string)
//
// Returns: IList[ILocalizedTextObject] containing localized text entries, or nil if reading fails
func ReadDataListWithIlist(filename string, languageCode string, creator func([]byte, []byte, int, string) components.ILocalizedTextObject) components.IList[components.ILocalizedTextObject] {
	fileAccessor, err := common.NewFileAccessor(filename)
	if err != nil {
		common.LogVerbose("Error accessing file: %v", err)
		return nil
	}

	if !fileAccessor.Exists {
		common.LogVerbose("File does not exist: %s", filename)
		return nil
	}

	data, err := os.ReadFile(fileAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading file: %v", err)
		return nil
	}

	if len(data) < 16 { // Minimum header size
		common.LogVerbose("File too small: %s", filename)
		return nil
	}

	offset := 8

	minIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	maxIndex := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Read individual length and total length
	individualLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2
	totalLength := int(binary.LittleEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// Skip 4 bytes
	offset += 4

	// Verify we have enough data
	if offset+totalLength > len(data) {
		common.LogVerbose("Insufficient data in file: %s", filename)
		return nil
	}

	// Read data bytes
	dataBytes := data[offset : offset+totalLength]
	offset += totalLength

	// Read string bytes (remaining data)
	stringBytes := data[offset:]

	// Create IList with pre-calculated capacity for optimal memory allocation
	count := maxIndex - minIndex
	objectCount := count + 1
	objects := components.NewList[components.ILocalizedTextObject](objectCount)

	for i := 0; i <= count; i++ {
		from := i * individualLength
		to := (i + 1) * individualLength

		if to > len(dataBytes) {
			break
		}
		objData := dataBytes[from:to]
		obj := creator(objData, stringBytes, individualLength, languageCode)
		objects.Add(obj)

		if common.IsVerboseMode() {
			offsetStr := fmt.Sprintf("%04X", (i*individualLength)+0x14)
			indexStr := fmt.Sprintf("%d", i+minIndex)
			fmt.Printf("%s (Offset %s) - %s\n", indexStr, offsetStr, obj.ToString(languageCode))
		}
	}

	return objects
}

// ReadAllEventFiles reads all event files from the specified events directory and populates
// the global EVENTS data structure. This function handles the complete event loading process,
// including directory traversal, file discovery, and localization loading.
//
// The function performs the following operations:
// 1. Validates the events directory exists and is accessible
// 2. Discovers all event subdirectories and files using IList for efficient collection handling
// 3. Reads each event file with full localization support
// 4. Populates the components.EVENTS global map
//
// Directory structure expected:
// - events/
//   - xx/ (two-letter prefixes)
//   - eventID/ (event directories)
//   - eventID.ebp (event binary file)
//
// Parameters:
//   - eventsFolder: FileAccessor pointing to the root events directory
//
// Returns: error if directory access fails or critical errors occur during processing
func ReadAllEventFiles(eventsFolder common.FileAccessor) error {
	if !eventsFolder.Exists {
		if common.IsVerboseMode() {
			fmt.Println("Cannot locate events at:", eventsFolder)
		}
		return fmt.Errorf("events directory not found: %s", eventsFolder.ResolvedPath)
	}

	eventIDs, err := discoverEventFiles(eventsFolder)
	if err != nil {
		return fmt.Errorf("failed to discover event files: %w", err)
	}

	return loadDiscoveredEvents(eventIDs)
}

// discoverEventFiles discovers all event files in the events directory structure.
// This function traverses the two-level directory structure used by the game's event system
// and builds an IList containing all available event IDs.
//
// Parameters:
//   - eventsFolder: FileAccessor pointing to the root events directory
//
// Returns: IList[string] containing discovered event IDs, or error if directory traversal fails
func discoverEventFiles(eventsFolder common.FileAccessor) (components.IList[string], error) {
	entries, err := os.ReadDir(eventsFolder.ResolvedPath)
	if err != nil {
		common.LogVerbose("Error reading events directory: %s", eventsFolder.ResolvedPath)
		return nil, fmt.Errorf("failed to read events directory: %w", err)
	}

	eventFiles := components.NewList[string](len(entries))

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip blitzball events if requested
		if common.SkipBlitzballEvents && entry.Name() == "bl" {
			continue
		}

		subEventIDs, err := discoverSubdirectoryEvents(eventsFolder.ResolvedPath, entry.Name())
		if err != nil {
			common.LogVerbose("Warning: failed to read subdirectory %s: %v", entry.Name(), err)
			continue
		}

		eventFiles.AddAll(subEventIDs.GetItems())
	}

	sort.Strings(eventFiles.GetItems())
	return eventFiles, nil
}

// discoverSubdirectoryEvents discovers event files within a specific subdirectory.
// This handles the second level of the event directory structure and returns the results
// as an IList[string] for type safety and efficient collection operations.
//
// Parameters:
//   - eventsRoot: Root events directory path
//   - subdirName: Name of the subdirectory to scan
//
// Returns: IList[string] containing event IDs found in the subdirectory, or error if access fails
func discoverSubdirectoryEvents(eventsRoot, subdirName string) (components.IList[string], error) {
	subPath := filepath.Join(eventsRoot, subdirName)
	subEntries, err := os.ReadDir(subPath)
	if err != nil {
		return nil, err
	}

	eventIDs := components.NewList[string](len(subEntries))
	for _, subEntry := range subEntries {
		if !subEntry.IsDir() || strings.HasPrefix(subEntry.Name(), ".") {
			continue
		}
		eventIDs.Add(subEntry.Name())
	}

	return eventIDs, nil
}

// loadDiscoveredEvents loads all discovered event files and populates the global EVENTS map.
// This function processes each event ID using IList.ForEach for efficient iteration and
// handles loading failures gracefully without interrupting the overall loading process.
//
// Parameters:
//   - eventIDs: IList[string] containing event IDs to load
//
// Returns: error only if critical system-level failures occur
func loadDiscoveredEvents(eventIDs components.IList[string]) error {
	eventIDs.ForEach(func(eventID string) {
		eventFile, err := ReadCompleteEventFile(eventID)
		if err != nil {
			fmt.Printf("failed to read event file %s: %v\n", eventID, err)
			return
		}
		if eventFile != nil {
			components.EVENTS[eventID] = eventFile
		}
	})
	return nil
}

// ReadCompleteEventFile reads a complete event file with all localizations.
// This function handles the full event loading process including binary file reading,
// localization loading, and validation.
//
// The function performs these operations:
// 1. Validates the event ID format
// 2. Constructs file paths using the standard event directory structure
// 3. Reads the main event binary file (.ebp)
// 4. Loads localized string files (.bin) for all supported languages
// 5. Combines all data into a complete EventFile object
//
// Parameters:
//   - eventID: Unique identifier for the event (e.g., "ev001", "btl_001")
//
// Returns: Complete EventFile with all localizations, or nil/error if loading fails
func ReadCompleteEventFile(eventID string) (*components.EventFile, error) {
	if len(eventID) < 2 {
		if common.IsVerboseMode() {
			fmt.Printf("Invalid event ID: %s\n", eventID)
		}
		return nil, fmt.Errorf("invalid event ID: %s", eventID)
	}

	// Handle special cases for specific game versions
	if common.GetGameVersionString() == "ffx2" && eventID == "crcr0000" {
		if common.IsVerboseMode() {
			fmt.Println("Skipping crcr0000 event in FFX-2")
		}
		return nil, nil
	}

	pathInfo := buildEventFilePaths(eventID)

	originalsPath, err := common.NewFileAccessor(pathInfo.EventFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve event file path: %w", err)
	}

	if !originalsPath.Exists || originalsPath.Size == 0 {
		common.LogVerbose("Skipping event due to missing or empty original file: %s", originalsPath)
		return nil, nil
	}

	eventFile, err := ReadEventBinaryFile(eventID, originalsPath)
	if err != nil || eventFile == nil {
		return nil, fmt.Errorf("failed to read event file %s: %w", eventID, err)
	}

	localizedStrings := components.ReadLocalizedStringFiles(pathInfo.LocalizationPattern)
	if localizedStrings != nil {
		eventFile.AddLocalizations(localizedStrings)
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Event content preview: %s", eventFile.String())
	}

	return eventFile, nil
}

type EventFilePaths struct {
	EventID             string
	Shortened           string
	MidPath             string
	EventFilePath       string
	LocalizationPattern string
}

// buildEventFilePaths constructs all necessary file paths for an event based on its ID.
// This function encapsulates the standard path construction logic used throughout
// the event system.
//
// Parameters:
//   - eventID: The event identifier
//
// Returns: EventFilePaths structure with all constructed paths
func buildEventFilePaths(eventID string) EventFilePaths {
	shortened := eventID[:2]
	midPath := filepath.Join(shortened, eventID, eventID)

	return EventFilePaths{
		EventID:             eventID,
		Shortened:           shortened,
		MidPath:             midPath,
		EventFilePath:       filepath.Join(common.GetPathOriginalsEvent(), midPath+".ebp"),
		LocalizationPattern: "event/obj_ps3/" + midPath + ".bin",
	}
}

// ReadEventBinaryFile reads a single event binary file and creates an EventFile object.
// This function handles the low-level binary file reading and EventFile object creation.
//
// The function performs these operations:
// 1. Validates file existence and accessibility
// 2. Reads the complete binary file content
// 3. Creates and initializes an EventFile object
// 4. Performs basic validation on the loaded data
//
// Parameters:
//   - eventID: Unique identifier for the event
//   - pathAccessor: FileAccessor pointing to the event binary file
//
// Returns: EventFile object with binary data loaded, or error if reading fails
func ReadEventBinaryFile(eventID string, pathAccessor common.FileAccessor) (*components.EventFile, error) {
	if !pathAccessor.Exists {
		return nil, fmt.Errorf("event file not found: %s", pathAccessor.ResolvedPath)
	}

	data, err := common.ReadFile(pathAccessor.ResolvedPath)
	if err != nil {
		common.LogVerbose("Failed to read event file: %s", pathAccessor.ResolvedPath)
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data read from event file: %s", pathAccessor.ResolvedPath)
	}

	eventFile := components.NewEventFile(eventID, data)

	common.LogVerbose("Successfully read event file: %s (%d bytes)", eventID, len(data))
	return eventFile, nil
}
