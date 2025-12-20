package exporters

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	"fmt"
	"path/filepath"
	"sort"
)

// createNameDescriptionJSON creates a JSON file with name and description data.
//
// Parameters:
//   - nameDescriptionList: Slice of NameDescriptionData structures to export
//   - fileName: Name of the JSON file to create
//
// Returns: error if file creation fails
func createNameDescriptionJSON(nameDescriptionList []*NameDescriptionData, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, fileName)
	if err := common.SaveAsJSON(nameDescriptionList, jsonPath); err != nil {
		return fmt.Errorf("error saving JSON data to file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}

// createNameOnlyJSON creates a JSON file with name-only data.
//
// Parameters:
//   - data: Slice of NameOnlyData structures to export
//   - fileName: Name of the JSON file to create
//
// Returns: error if file creation fails
func createNameOnlyJSON(data []*NameOnlyData, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, fileName)
	if err := common.SaveAsJSON(data, jsonPath); err != nil {
		return fmt.Errorf("error saving JSON data to file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-only data to JSON: %s", jsonPath)
	return nil
}

// serializeNameDescriptionToJSON converts IList data to JSON format for name-description objects.
//
// Parameters:
//   - objects: IList containing LocalizedTextObject entries with name and description
//   - jsonFileName: Name of the output JSON file
//
// Returns: error if serialization fails
func serializeNameDescriptionToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var nameDescData []*NameDescriptionData

	objects.RangeIndex(func(i int, nameDescObj datastore.IGlobalLocalizedTextObject) {
		if nameDescObj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &NameDescriptionData{
			NameOnlyData: NameOnlyData{
				ID:   i,
				Name: make(map[string]string),
			},
			Description: make(map[string]string),
		}

		nameKeyed := nameDescObj.GetKeyedString("name")
		descKeyed := nameDescObj.GetKeyedString("description")

		for locKey := range common.SupportedLanguages {
			if nameKeyed != nil {
				nameText := nameKeyed.GetLocalizedString(locKey)
				if nameText != "" {
					data.Name[locKey] = nameText
				}
			}

			if descKeyed != nil {
				descText := descKeyed.GetLocalizedString(locKey)
				if descText != "" {
					data.Description[locKey] = descText
				}
			}
		}

		if len(data.Name) > 0 || len(data.Description) > 0 {
			nameDescData = append(nameDescData, data)
		}
	})

	return createNameDescriptionJSON(nameDescData, jsonFileName)
}

// serializeNameOnlyToJSON converts IList data to JSON format for name-only objects.
//
// Parameters:
//   - objects: IList containing LocalizedTextObject entries with name only
//   - jsonFileName: Name of the output JSON file
//
// Returns: error if serialization fails
func serializeNameOnlyToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var nameOnlyData []*NameOnlyData

	objects.RangeIndex(func(i int, nameObj datastore.IGlobalLocalizedTextObject) {
		if nameObj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &NameOnlyData{
			ID:   i,
			Name: make(map[string]string),
		}

		nameKeyed := nameObj.GetKeyedString("name")

		for locKey := range common.SupportedLanguages {
			if nameKeyed != nil {
				nameText := nameKeyed.GetLocalizedString(locKey)
				if nameText != "" {
					data.Name[locKey] = nameText
				}
			}
		}

		if len(data.Name) > 0 {
			nameOnlyData = append(nameOnlyData, data)
		}
	})

	return createNameOnlyJSON(nameOnlyData, jsonFileName)
}

// ExportKeyItemsToJSON exports key items data from objectsfile.KEY_ITEMS to a JSON file.
//
// This function reads the key items data loaded in memory and exports it to
// "key_items_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportKeyItemsToJSON() error {
	keyItems := datastore.KeyItems
	if keyItems == nil || keyItems.IsEmpty() {
		return fmt.Errorf("KEY_ITEMS data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(keyItems, "key_items_all_localizations.json")
}

// ExportCommandsToJSON exports commands data from objectsfile.COMMANDS to a JSON file.
//
// This function reads the commands data loaded in memory and exports it to
// "commands_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportCommandsToJSON() error {
	if datastore.Commands.IsEmpty() {
		return fmt.Errorf("COMMANDS data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(datastore.Commands, "commands_all_localizations.json")
}

// ExportItemsToJSON exports items data from objectsfile.ITEMS to a JSON file.
//
// This function reads the items data loaded in memory and exports it to
// "items_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportItemsToJSON() error {
	if datastore.Items.IsEmpty() {
		return fmt.Errorf("ITEMS data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(datastore.Items, "items_all_localizations.json")
}

// ExportArmsToJSON exports arms data from objectsfile.ARMS_TEXT to a JSON file.
//
// This function reads the arms text data loaded in memory and exports it to
// "arms_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportArmsToJSON() error {
	if datastore.ArmsTxt.IsEmpty() {
		return fmt.Errorf("ARMS_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(datastore.ArmsTxt, "arms_all_localizations.json")
}

// ExportConfigToJSON exports config data from objectsfile.CONFIG_TEXT to a JSON file.
//
// This function reads the config text data loaded in memory and exports it to
// "config_text_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportConfigToJSON() error {
	if objectsfile.CONFIG_TEXT == nil || objectsfile.CONFIG_TEXT.IsEmpty() {
		return fmt.Errorf("CONFIG_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.CONFIG_TEXT, "config_text_all_localizations.json")
}

// ExportItemCommandsToJSON exports item descriptions from objectsfile.ITEM_TEXT to a JSON file.
//
// This function reads the item text data loaded in memory and exports it to
// "item_commands_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportItemCommandsToJSON() error {
	if objectsfile.ITEM_TEXT == nil || objectsfile.ITEM_TEXT.IsEmpty() {
		return fmt.Errorf("ITEM_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.ITEM_TEXT, "item_commands_all_localizations.json")
}

// ExportMainMenuToJSON exports main menu data from objectsfile.MMAIN_TEXT to a JSON file.
//
// This function reads the main menu text data loaded in memory and exports it to
// "main_menu_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportMainMenuToJSON() error {
	if objectsfile.MMAIN_TEXT == nil || objectsfile.MMAIN_TEXT.IsEmpty() {
		return fmt.Errorf("MMAIN_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.MMAIN_TEXT, "main_menu_all_localizations.json")
}

// ExportPlayerRoomToJSON exports player room data from objectsfile.PLAYER_ROOM to a JSON file.
//
// This function reads the player room text data loaded in memory and exports it to
// "player_room_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportPlayerRoomToJSON() error {
	if objectsfile.PLAYER_ROOM == nil || objectsfile.PLAYER_ROOM.IsEmpty() {
		return fmt.Errorf("PLAYER_ROOM data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.PLAYER_ROOM, "player_room_all_localizations.json")
}

// ExportBuildToJSON exports build data from objectsfile.BUILD_TEXT to a JSON file.
//
// This function reads the build text data loaded in memory and exports it to
// "build_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportBuildToJSON() error {
	if objectsfile.BUILD_TEXT == nil || objectsfile.BUILD_TEXT.IsEmpty() {
		return fmt.Errorf("BUILD_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(objectsfile.BUILD_TEXT, "build_all_localizations.json")
}

// ExportBattleToJSON exports battle data from objectsfile.BTL_TEXT to a JSON file.
//
// This function reads the battle text data loaded in memory and exports it to
// "battle_text_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportBattleToJSON() error {
	if datastore.BattleTxt.IsEmpty() {
		return fmt.Errorf("BTL_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(datastore.BattleTxt, "battle_text_all_localizations.json")
}

// ExportBattleEndToJSON exports battle end data from objectsfile.BTLEND_TEXT to a JSON file.
//
// This function reads the battle end text data loaded in memory and exports it to
// "battle_end_text_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportBattleEndToJSON() error {
	if datastore.BattleEndTxt.IsEmpty() {
		return fmt.Errorf("BTLEND_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(datastore.BattleEndTxt, "battle_end_text_all_localizations.json")
}

// ExportMonsterMagic1ToJSON exports monster magic 1 data from objectsfile.MONMAGIC1 to a JSON file.
//
// This function reads the monster magic 1 data loaded in memory and exports it to
// "monster_magic1_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportMonsterMagic1ToJSON() error {
	if objectsfile.MONMAGIC1 == nil || objectsfile.MONMAGIC1.IsEmpty() {
		return fmt.Errorf("MONMAGIC1 data not loaded or empty")
	}

	return serializeNameOnlyToJSON(objectsfile.MONMAGIC1, "monster_magic1_all_localizations.json")
}

// ExportMonsterMagic2ToJSON exports monster magic 2 data from objectsfile.MONMAGIC2 to a JSON file.
//
// This function reads the monster magic 2 data loaded in memory and exports it to
// "monster_magic2_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportMonsterMagic2ToJSON() error {
	if objectsfile.MONMAGIC2 == nil || objectsfile.MONMAGIC2.IsEmpty() {
		return fmt.Errorf("MONMAGIC2 data not loaded or empty")
	}

	return serializeNameOnlyToJSON(objectsfile.MONMAGIC2, "monster_magic2_all_localizations.json")
}

// ExportNameToJSON exports name data from objectsfile.NAME_TEXT to a JSON file.
//
// This function reads the name text data loaded in memory and exports it to
// "names_all_localizations.json" with all available localizations.
//
// Returns: error if export fails or data is not loaded
func ExportNameToJSON() error {
	if objectsfile.NAME_TEXT == nil || objectsfile.NAME_TEXT.IsEmpty() {
		return fmt.Errorf("NAME_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(objectsfile.NAME_TEXT, "names_all_localizations.json")
}

// getLocalizationKeys returns all available localization keys
// This function returns the localization keys from the common package
func getLocalizationKeys() []string {
	var keys []string
	for key := range common.SupportedLanguages {
		keys = append(keys, key)
	}
	return keys
}

// getSortedLocalizationKeys returns localization keys sorted alphabetically
func getSortedLocalizationKeys() []string {
	localizationKeys := getLocalizationKeys()
	sort.Strings(localizationKeys)
	return localizationKeys
}

// buildEventStringData creates EventStringData from an event string with all localizations
//
// Parameters:
//   - index: Index of the string in the event
//   - str: String object that implements GetLocalizedString method
//   - localizationKeys: List of localization keys to process
//
// Returns: EventStringData with index and all localized text
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

// processEventFromMemory processes an event from memory and creates EventFileData
//
// Parameters:
//   - eventID: ID of the event to process
//   - localizationKeys: List of localization keys to include
//
// Returns: EventFileData pointer or nil if no valid data found
func processEventFromMemory(eventID string, localizationKeys []string) *EventFileData {
	eventFile := event.GetEvent(eventID)
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

// processEventFromFile processes an event from file and creates EventFileData
//
// Parameters:
//   - eventID: ID of the event to process
//   - localizationKeys: List of localization keys to include
//
// Returns: EventFileData pointer or nil if no valid data found
func processEventFromFile(eventID string, localizationKeys []string) *EventFileData {
	if common.IsVerboseMode() {
		common.LogVerbose("Exporting event file to JSON: %s", eventID)
	}

	eventFileStrings, err := event.ReadLocalizedEventStrings(eventID)
	if err != nil {
		common.LogVerbose("Error loading localized strings: %v", err)
		return nil
	}

	if len(eventFileStrings) == 0 {
		return nil
	}

	eventData := EventFileData{
		ID:      eventID,
		Strings: make([]EventStringData, 0, len(eventFileStrings)),
	}

	for i, str := range eventFileStrings {
		stringData := buildEventStringData(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	return &eventData
}

// writeEventJSONFile writes event data to a JSON file
//
// Parameters:
//   - events: Slice of EventFileData to write
//   - fileName: Name of the JSON file to create
//
// Returns: error if file creation fails
func writeEventJSONFile(events []EventFileData, fileName string) error {
	if len(events) == 0 {
		common.LogVerbose("No events with string data found to export to JSON")
		return nil
	}

	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	filePath := filepath.Join(editsPath, fileName)
	if err := common.SaveAsJSON(events, filePath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}

	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: %d", len(events))
	return nil
}

// ExportAllEventsToJSON exports all event data to a JSON file with all localizations.
//
// This function processes all event entries and exports them to
// "events_all_localizations.json" with all available localizations.
// Only exports events that have string data (skips empty events).
//
// Returns: error if export fails or data is not loaded
func ExportAllEventsToJSON() error {
	fileName := "events_all_localizations.json"
	localizationKeys := getSortedLocalizationKeys()
	eventIDs := event.GetAllEventIDs()

	var allEvents []EventFileData
	var count int

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			common.LogVerbose("Skipping event %s: no strings found", eventID)
			continue
		}

		allEvents = append(allEvents, *eventData)
		count++
	}

	common.LogVerbose("Total events processed: %d from %d", count, len(eventIDs))
	return writeEventJSONFile(allEvents, fileName)
}

// ExportEventsForLocalizationToJSON exports event data for a specific language to a JSON file.
//
// This function processes all event entries and exports them to a JSON file
// named "events_{languageCode}.json" with the specified localization only.
//
// Parameters:
//   - languageCode: Language code for localization (e.g., "us", "jp")
//
// Returns: error if export fails or data is not loaded
func ExportEventsForLocalizationToJSON(languageCode string) error {
	eventIDs := event.GetAllEventIDs()
	localizationKeys := []string{languageCode}

	var allEvents []EventFileData

	for _, eventID := range eventIDs {
		eventData := processEventFromMemory(eventID, localizationKeys)
		if eventData == nil {
			continue
		}

		allEvents = append(allEvents, *eventData)
	}

	if len(allEvents) == 0 {
		return fmt.Errorf("no events with string data found for localization %s", languageCode)
	}

	fileName := fmt.Sprintf("events_%s.json", languageCode)
	return writeEventJSONFile(allEvents, fileName)
}

// ExportSingleEventToJSON exports a single event's data to a JSON file with all localizations.
//
// This function processes the event identified by eventId and exports it to
// "event_{eventId}_all_localizations.json" with all available localizations.
//
// Parameters:
//   - eventId: Unique identifier for the event to export
//
// Returns: error if export fails or data is not loaded
func ExportSingleEventToJSON(eventId string) error {
	localizationKeys := getSortedLocalizationKeys()

	eventData := processEventFromFile(eventId, localizationKeys)
	if eventData == nil {
		return fmt.Errorf("no data found for event %s", eventId)
	}

	allEvents := []EventFileData{*eventData}
	fileName := "event_" + eventId + "_all_localizations.json"

	return writeEventJSONFile(allEvents, fileName)
}
