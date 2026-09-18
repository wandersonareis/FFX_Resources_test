package event

import (
	"ffxresources/backend/common"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
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

func eventFileExport(eventData EventFileData, version common.GameVersion) models.EventFileExport {
	strings := make([]models.EventStringDataExport, 0, len(eventData.Strings))
	for _, s := range eventData.Strings {
		strings = append(strings, models.EventStringDataExport{Index: s.Index, Text: s.Text})
	}
	return models.EventFileExport{
		Metadata: models.NewEventFileInfo(eventData.ID, version),
		ID:       eventData.ID,
		Strings:  strings,
	}
}

func eventJSONPath(fileName string, version common.GameVersion) (string, error) {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return "", fmt.Errorf("error creating edits directory: %w", err)
	}
	return filepath.Join(editsPath, common.WithVersionSuffixFor(fileName, version)), nil
}

func writeSingleEventJSONFile(event EventFileData, fileName string, version common.GameVersion) error {
	filePath, err := eventJSONPath(fileName, version)
	if err != nil {
		return err
	}
	export := []models.EventFileExport{eventFileExport(event, version)}
	if err := models.SaveDataFile(export, filePath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}
	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: 1")
	return nil
}

func writeEventsJSONFile(events []EventFileData, fileName string, version common.GameVersion) error {
	filePath, err := eventJSONPath(fileName, version)
	if err != nil {
		return err
	}
	stringsMap := make(map[string]models.EventFileExport, len(events))
	for _, e := range events {
		stringsMap[e.ID] = eventFileExport(e, version)
	}
	export := models.EventsFileExport{Strings: stringsMap}
	if err := models.SaveDataFile(export, filePath); err != nil {
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

func exportSingleEventToJSON(version common.GameVersion, eventID string) error {
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
	return writeSingleEventJSONFile(*eventData, fileName, version)
}

func exportEventsToJSON(version common.GameVersion, fileName string, eventIDs []string, localizationKeys []string) error {
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
	return writeEventsJSONFile(allEvents, fileName, version)
}

func ExportAllEventsToJSON() error {
	return ExportAllEventsToJSONForVersion(currentGameVersion())
}

func ExportAllEventsToJSONForVersion(version common.GameVersion) error {
	eventIDs := GetAllEventIDs(version)
	localizationKeys := getSortedLocalizationKeys()
	return exportEventsToJSON(version, "events_all_localizations.json", eventIDs, localizationKeys)
}
