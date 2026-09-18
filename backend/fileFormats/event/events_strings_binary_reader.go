package event

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/models"
)

func ReadAllEventFiles(eventsFolder common.FileAccessor, version common.GameVersion) error {
	if !eventsFolder.Exists {
		return fmt.Errorf("events directory not found: %s", eventsFolder.ResolvedPath)
	}

	first := true

	loadedEvents := components.NewList[*models.EventFileInfo](0)
	for _, localization := range SortedSupportedLocalizations() {
		eventInfos, err := discoverEventFiles(eventsFolder, localization, version)
		if err != nil {
			if first {
				return fmt.Errorf("failed to discover event files: %w", err)
			}
			common.LogVerbose("failed to discover event files for %s: %v", localization, err)
			continue
		}

		for _, info := range eventInfos {
			if first {
				loadedEvents.Add(info)
				continue
			}
			if loadedEvents.TryAdd(info) {
				common.LogInfo("event id added from %s: %s", localization, info.EventID)
			}
		}

		first = false
	}

	for _, info := range loadedEvents.Items() {
		if _, err := ReadCompleteEventFile(info); err != nil {
			common.LogVerbose("failed to read event file %s: %v", info.EventID, err)
			continue
		}
	}

	common.LogInfo("events loaded: %d", loadedEvents.Len())
	return nil
}

func SortedSupportedLocalizations() []string {
	keys := make([]string, 0, len(common.SupportedLanguages))
	for loc := range common.SupportedLanguages {
		if loc == common.DefaultLocalization {
			continue
		}
		keys = append(keys, loc)
	}
	sort.Strings(keys)
	return append([]string{common.DefaultLocalization}, keys...)
}

func discoverEventFiles(eventsFolder common.FileAccessor, localization string, version common.GameVersion) ([]*models.EventFileInfo, error) {
	entries, err := os.ReadDir(eventsFolder.ResolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read events directory: %w", err)
	}

	var eventInfos []*models.EventFileInfo
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if common.SkipBlitzballEvents && entry.Name() == "bl" {
			continue
		}

		subInfos, err := discoverSubdirectoryEvents(eventsFolder.ResolvedPath, entry.Name(), version)
		if err != nil {
			common.LogVerbose("failed to read subdirectory %s: %v", entry.Name(), err)
			continue
		}
		eventInfos = append(eventInfos, subInfos...)
	}

	sort.Slice(eventInfos, func(i, j int) bool {
		return eventInfos[i].EventID < eventInfos[j].EventID
	})
	return eventInfos, nil
}

func discoverSubdirectoryEvents(eventsRoot, subdirName string, version common.GameVersion) ([]*models.EventFileInfo, error) {
	subPath := filepath.Join(eventsRoot, subdirName)
	subEntries, err := os.ReadDir(subPath)
	if err != nil {
		return nil, err
	}

	var eventInfos []*models.EventFileInfo
	for _, subEntry := range subEntries {
		if !subEntry.IsDir() || strings.HasPrefix(subEntry.Name(), ".") {
			continue
		}
		info := models.NewEventFileInfo(subEntry.Name(), version)
		if info != nil {
			eventInfos = append(eventInfos, info)
		}
	}

	return eventInfos, nil
}

func ReadCompleteEventFile(info *models.EventFileInfo) (*EventFile, error) {
	if info == nil || len(info.EventID) < 2 {
		return nil, fmt.Errorf("invalid event ID")
	}

	if info.Version == common.GameVersionFFX2 && info.EventID == "crcr0000" {
		return nil, nil
	}

	localizedStrings := ReadLocalizedStringFiles(info.LocalizationPattern, info.Version)
	if len(localizedStrings) == 0 {
		common.LogVerbose("Skipping event due to missing or empty localization: %s", info.EventID)
		return nil, nil
	}

	eventFile := &EventFile{
		ID:      info.EventID,
		Version: info.Version,
		Strings: localizedStrings,
	}

	SetEvent(info.Version, info.EventID, eventFile)
	return eventFile, nil
}

