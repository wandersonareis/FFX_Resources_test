package event

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"fmt"
	"path/filepath"
	"sort"
)

func currentGameVersion() common.GameVersion {
	return interactions.CurrentGameVersion()
}

func getLocalizationKeys() []string {
	var keys []string
	for key := range common.SupportedLanguages {
		keys = append(keys, key)
	}
	return keys
}

func getSortedLocalizationKeys() []string {
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)
	return localizationKeys
}

func buildEventStringData(index int, str interface{ GetLocalizedString(string) string }, localizationKeys []string) EventStringData {
	stringData := EventStringData{
		Index: index,
		Text:  make(map[string]string),
	}
	for _, langKey := range localizationKeys {
		value := str.GetLocalizedString(langKey)
		stringData.Text[langKey] = value
	}
	return stringData
}

func eventJSONPath(fileName string, version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor(fileName, version)), nil
}

func writeSingleEventJSONFile(event EventFileData, fileName string, version common.GameVersion, formatter IEventsFormatter) error {
	filePath, err := eventJSONPath(fileName, version)
	if err != nil {
		return err
	}
	export := []EventFileData{event}
	raw, err := formatter.Marshal(export, version)
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", filePath, err)
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}
	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: 1")
	return nil
}

func writeEventsJSONFile(events []EventFileData, fileName string, version common.GameVersion, formatter IEventsFormatter) error {
	filePath, err := eventJSONPath(fileName, version)
	if err != nil {
		return err
	}
	raw, err := formatter.Marshal(events, version)
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", filePath, err)
	}
	if err := common.WriteBytesToFile(filePath, raw); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}
	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: %d", len(events))
	return nil
}

func processEventFromMemoryForVersion(version common.GameVersion, eventID string, localizationKeys []string) *EventFileData {
	eventFile := GetEvent(version, eventID)
	if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
		return nil
	}
	eventData := EventFileData{
		ID:      eventFile.ID,
		Strings: make([]EventStringData, 0, len(eventFile.Strings)),
	}
	for i, str := range eventFile.Strings {
		stringData := buildEventStringData(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}
	if len(eventData.Strings) == 0 {
		common.LogVerbose("No strings found for event %s, skipping", eventID)
		return nil
	}
	return &eventData
}

func exportSingleEventToJSON(version common.GameVersion, eventID string, formatter IEventsFormatter) error {
	eventIDs := GetAllEventIDs(version)
	if len(eventIDs) == 0 {
		return fmt.Errorf("no events loaded for version %s", version)
	}
	localizationKeys := getSortedLocalizationKeys()
	eventData := processEventFromMemoryForVersion(version, eventID, localizationKeys)
	if eventData == nil {
		return fmt.Errorf("no data found for event %s", eventID)
	}
	fileName := "event_" + eventID + "_all_localizations.json"
	return writeSingleEventJSONFile(*eventData, fileName, version, formatter)
}

func exportEventsToJSON(version common.GameVersion, fileName string, eventIDs []string, localizationKeys []string, formatter IEventsFormatter) error {
	var allEvents []EventFileData
	for _, eventID := range eventIDs {
		eventData := processEventFromMemoryForVersion(version, eventID, localizationKeys)
		if eventData == nil {
			continue
		}
		allEvents = append(allEvents, *eventData)
	}
	if len(allEvents) == 0 {
		return fmt.Errorf("no events with string data found for localization")
	}
	return writeEventsJSONFile(allEvents, fileName, version, formatter)
}

func ExportAllEventsToJSON(formatter IEventsFormatter) error {
	return ExportAllEventsToJSONForVersion(currentGameVersion(), formatter)
}

func ExportAllEventsToJSONForVersion(version common.GameVersion, formatter IEventsFormatter) error {
	eventIDs := GetAllEventIDs(version)
	sortedIDs := make([]string, len(eventIDs))
	copy(sortedIDs, eventIDs)
	sort.Strings(sortedIDs)
	localizationKeys := getSortedLocalizationKeys()
	return exportEventsToJSON(version, "events_all_localizations.json", sortedIDs, localizationKeys, formatter)
}
