package event

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReadAllEventFiles reads all event files from the specified events directory and registers
// them in the datastore. This function handles the complete event loading process,
// including directory traversal, file discovery, and localization loading.
//
// The function performs the following operations:
// 1. Validates the events directory exists and is accessible
// 2. Discovers all event subdirectories and files using IList for efficient collection handling
// 3. Reads each event file with full localization support
// 4. Registers events directly in the datastore as the single source of truth
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

		eventFiles.AddAll(subEventIDs.Items())
	}

	sort.Strings(eventFiles.Items())
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

// loadDiscoveredEvents loads all discovered event files and registers them directly in the datastore.
// This function processes each event ID using IList.Range for efficient iteration and
// handles loading failures gracefully without interrupting the overall loading process.
// All events are registered directly in the datastore as the single source of truth.
//
// Parameters:
//   - eventIDs: IList[string] containing event IDs to load
//
// Returns: error only if critical system-level failures occur
func loadDiscoveredEvents(eventIDs components.IList[string]) error {
	eventIDs.Range(func(eventID string) {
		eventFile, err := ReadCompleteEventFile(eventID)
		if err != nil {
			fmt.Printf("failed to read event file %s: %v\n", eventID, err)
			return
		}
		if eventFile != nil {
			// Registrar diretamente no datastore (fonte única da verdade)
			SetEvent(eventID, eventFile)
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
func ReadCompleteEventFile(eventID string) (*EventFile, error) {
	if len(eventID) < 2 {
		if common.IsVerboseMode() {
			fmt.Printf("Invalid event ID: %s\n", eventID)
		}
		return nil, fmt.Errorf("invalid event ID: %s", eventID)
	}

	// TODO: find better solution for this junk event files
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

	event, err := ReadEventBinaryFile(eventID, originalsPath)
	if err != nil || event == nil {
		return nil, fmt.Errorf("failed to read event file %s: %w", eventID, err)
	}

	localizedStrings := ReadLocalizedStringFiles(pathInfo.LocalizationPattern)
	if localizedStrings != nil {
		event.AddLocalizations(localizedStrings)
	}

	if common.IsVerboseMode() {
		common.LogVerbose("Event content preview: %s", event.String())
	}

	return event, nil
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
func ReadEventBinaryFile(eventID string, pathAccessor common.FileAccessor) (*EventFile, error) {
	if !pathAccessor.Exists {
		return nil, fmt.Errorf("event file not found: %s", pathAccessor.ResolvedPath)
	}

	data := pathAccessor.ReadBytes()
	if data == nil {
		return nil, fmt.Errorf("failed to read event file: %v", pathAccessor)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data read from event file: %v", pathAccessor)
	}

	eventFile := NewEventFile(eventID, data)

	common.LogVerbose("Successfully read event file: %s (%d bytes)", eventID, len(data))
	return eventFile, nil
}
