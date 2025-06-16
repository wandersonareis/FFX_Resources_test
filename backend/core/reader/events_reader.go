package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ReadAllEvents(eventsFolder common.FileAccessor) error {
	if !eventsFolder.Exists {
		if common.IsVerboseMode() {
			fmt.Println("Cannot locate events at:", eventsFolder)
		}
		return fmt.Errorf("events directory not found: %s", eventsFolder.ResolvedPath)
	}

	entries, err := os.ReadDir(eventsFolder.ResolvedPath)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Println("Cannot list events:", err)
		}
		return fmt.Errorf("failed to read events directory: %w", err)
	}

	var eventFiles []string

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		// Skip blitzball events if requested
		if common.SkipBlitzballEvents && entry.Name() == "bl" {
			continue
		}

		subPath := filepath.Join(eventsFolder.ResolvedPath, entry.Name())
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
		eventFile, err := ReadEventFull(eventId)
		if err != nil {
			fmt.Printf("failed to read event file %s: %v\n", eventId, err)
			continue
		}
		components.EVENTS[eventId] = eventFile

	}
	return nil
}

func ReadEventFull(eventId string) (*components.EventFile, error) {
	if len(eventId) < 2 {
		if common.IsVerboseMode() {
			fmt.Printf("Invalid event ID: %s\n", eventId)
		}
		return nil, fmt.Errorf("invalid event ID: %s", eventId)
	}

	if common.GetGameVersionString() == "ffx2" && eventId == "crcr0000" {
		if common.IsVerboseMode() {
			fmt.Println("Skipping crcr0000 event in FFX-2")
		}
		return nil, nil
	}

	shortened := eventId[:2]

	midPath := filepath.Join(shortened, eventId, eventId)
	eventPath := filepath.Join(common.GetPathOriginalsEvent(), midPath+".ebp")
	originalsPath, err := common.NewFileAccessor(eventPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve event file path: %w", err)
	}

	if !originalsPath.Exists || originalsPath.Size == 0 {
		if common.IsVerboseMode() {
			fmt.Println("Skipping event due to missing or empty original file:", originalsPath)
		}
		return nil, nil
	}

	eventFile, err := ReadEventFile(eventId, originalsPath)
	if err != nil || eventFile == nil {
		return nil, fmt.Errorf("failed to read event file %s: %w", eventId, err)
	}

	localizedStrings := components.ReadLocalizedStringFiles("event/obj_ps3/" + midPath + ".bin")
	if localizedStrings != nil {
		eventFile.AddLocalizations(localizedStrings)
	}

	if common.IsVerboseMode() {
		textOutputPath := common.PathTextOutputRoot + "event/obj/" + shortened + "/" + eventId + ".txt"
		eventFileString := eventFile.String()

		if err := common.EnsurePathExists(textOutputPath); err != nil {
			return nil, fmt.Errorf("failed to ensure output directory exists: %w", err)
		}

		if common.IsVerboseMode() {
			fmt.Printf("Processed event %s, output would go to: %s\n", eventId, textOutputPath)
			fmt.Printf("Event content preview: %s\n", eventFileString)
		}
	}

	return eventFile, nil
}

func ReadEventFile(eventId string, path common.FileAccessor) (*components.EventFile, error) {
	if !path.Exists {
		return nil, fmt.Errorf("event file not found: %s", path.ResolvedPath)
	}

	data, err := common.ReadFile(path.ResolvedPath)
	if err != nil {
		if common.IsVerboseMode() {
			fmt.Printf("Error reading event file %s: %v\n", path.ResolvedPath, err)
		}
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("no data read from event file: %s", path.ResolvedPath)
	}

	eventFile := components.NewEventFile(eventId, data)

	if common.IsVerboseMode() {
		fmt.Printf("Successfully read event file: %s (%d bytes)\n", eventId, len(data))
	}

	return eventFile, nil
}
