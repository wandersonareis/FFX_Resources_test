package objectsfile

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/models"
)

// EventFileDataJSON represents an event file with all its strings for JSON export
type EventFileDataJSON struct {
	ID      string                `json:"id"`
	Strings []EventStringDataJSON `json:"strings"`
}

// EventStringDataJSON represents a single event string with its localizations
type EventStringDataJSON struct {
	Index int               `json:"index"`
	Text  map[string]string `json:"text"`
}

// objectsFilePatternPaths maps each exported objectsfile JSON to the pattern path of
// the original game binary it was extracted from, so its metadata can be reconstructed.
var objectsFilePatternPaths = map[string]string{
	"key_items_all_localizations.json":       "battle/kernel/important.bin",
	"commands_all_localizations.json":        "battle/kernel/command.bin",
	"items_all_localizations.json":           "battle/kernel/item.bin",
	"arms_all_localizations.json":            "battle/kernel/arms_txt.bin",
	"config_text_all_localizations.json":     "battle/kernel/config_txt.bin",
	"item_commands_all_localizations.json":   "battle/kernel/item_txt.bin",
	"main_menu_all_localizations.json":       "battle/kernel/mmain_txt.bin",
	"player_room_all_localizations.json":     "battle/kernel/ply_rom.bin",
	"battle_text_all_localizations.json":     "battle/kernel/btl_txt.bin",
	"battle_end_text_all_localizations.json": "battle/kernel/btlend_txt.bin",
	"monster_magic1_all_localizations.json":  "battle/kernel/monmagic1.bin",
	"monster_magic2_all_localizations.json":  "battle/kernel/monmagic2.bin",
	"build_all_localizations.json":           "battle/kernel/build_txt.bin",
	"names_all_localizations.json":           "battle/kernel/name_txt.bin",

	// FFX-2 (v2) kernel exclusivos.
	"a_ability_all_localizations.json":     "battle/kernel/a_ability.bin",
	"accessory_all_localizations.json":     "battle/kernel/accessory.bin",
	"job_all_localizations.json":           "battle/kernel/job.bin",
	"menu_text_all_localizations.json":     "battle/kernel/menu_txt.bin",
	"monster_magic_all_localizations.json": "battle/kernel/monmagic.bin",
	"monster_all_localizations.json":       "battle/kernel/monster.bin",
	"monster2_all_localizations.json":      "battle/kernel/monster2.bin",
	"oversoul_all_localizations.json":      "battle/kernel/oversoul.bin",
	"plate_all_localizations.json":         "battle/kernel/plate.bin",
	"player_save_all_localizations.json":   "battle/kernel/ply_save.bin",
	"save_text_all_localizations.json":     "battle/kernel/save_txt.bin",
}

// binaryMetadataForFile returns metadata describing the original binary for a given
// objectsfile export filename, or nil when the mapping is unknown.
func binaryMetadataForFile(fileName string) *models.FileMetadata {
	pattern, ok := objectsFilePatternPaths[fileName]
	if !ok {
		return nil
	}
	return models.NewFileMetadata(models.NewFileInfoFromPath(models.ObjectFileBinaryPath(pattern)))
}

func createJSON(dataObjectEntries []*JSONEntry, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	stringsBytes, err := json.Marshal(dataObjectEntries)
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", jsonPath, err)
	}

	export := models.ObjectsFileExport{
		Metadata: binaryMetadataForFile(fileName),
		Strings:  stringsBytes,
	}

	if err := models.SaveDataFile(export, jsonPath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}

// createNameDescriptionJSON creates a JSON file with name and description data.
func createNameDescriptionJSON(nameDescriptionList []*JSONEntry, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	stringsBytes, err := json.Marshal(nameDescriptionList)
	if err != nil {
		return fmt.Errorf("error marshaling data for %s: %w", jsonPath, err)
	}

	export := models.ObjectsFileExport{
		Metadata: binaryMetadataForFile(fileName),
		Strings:  stringsBytes,
	}

	if err := models.SaveDataFile(export, jsonPath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", jsonPath, err)
	}

	common.LogVerbose("Exported name-description data to JSON: %s", jsonPath)
	return nil
}

func ExportToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var jsonEntries []*JSONEntry

	objects.RangeIndex(func(i int, obj datastore.IGlobalLocalizedTextObject) {
		if obj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &JSONEntry{
			ID: i,
		}

		nameKeyed := obj.GetKeyedString("name")
		simplifiedNameKeyed := obj.GetKeyedString("simplifiedName")
		descKeyed := obj.GetKeyedString("description")
		simplifiedDescKeyed := obj.GetKeyedString("simplifiedDescription")
		effKeyed := obj.GetKeyedString("effect")

		if nameKeyed != nil {
			data.Name = make(map[string]string)
		}

		var abKeyed []datastore.IGlobalLocalizedKeyedStringObject
		for idx := 1; ; idx++ {
			ability := obj.GetKeyedString(fmt.Sprintf("ability%d", idx))
			if ability == nil {
				break
			}
			abKeyed = append(abKeyed, ability)
		}

		sensorKeyed := obj.GetKeyedString("sensorText")
		simplifiedSensorKeyed := obj.GetKeyedString("simplifiedSensorText")
		scanKeyed := obj.GetKeyedString("scanText")
		simplifiedScanKeyed := obj.GetKeyedString("simplifiedScanText")

		if len(abKeyed) > 0 {
			data.Abilities = make([]map[string]string, len(abKeyed))
			for idx := range data.Abilities {
				data.Abilities[idx] = make(map[string]string)
			}
		}

		for locKey := range common.SupportedLanguages {
			// Name
			if nameKeyed != nil {
				if nameText := nameKeyed.GetLocalizedString(locKey); nameText != "" {
					data.Name[locKey] = nameText
				}
			}

			// Description
			if descKeyed != nil {
				if descText := descKeyed.GetLocalizedString(locKey); descText != "" {
					if data.Description == nil {
						data.Description = make(map[string]string)
					}
					data.Description[locKey] = descText
				}
			}

			// SimplifiedName
			if simplifiedNameKeyed != nil {
				if simplifiedNameText := simplifiedNameKeyed.GetLocalizedString(locKey); simplifiedNameText != "" {
					if data.SimplifiedName == nil {
						data.SimplifiedName = make(map[string]string)
					}
					data.SimplifiedName[locKey] = simplifiedNameText
				}
			}

			// SimplifiedDescription
			if simplifiedDescKeyed != nil {
				if simplifiedDescText := simplifiedDescKeyed.GetLocalizedString(locKey); simplifiedDescText != "" {
					if data.SimplifiedDescription == nil {
						data.SimplifiedDescription = make(map[string]string)
					}
					data.SimplifiedDescription[locKey] = simplifiedDescText
				}
			}

			// Effect
			if effKeyed != nil {
				if effText := effKeyed.GetLocalizedString(locKey); effText != "" {
					if data.Effect == nil {
						data.Effect = make(map[string]string)
					}
					data.Effect[locKey] = effText
				}
			}

			// Abilities
			for idx, abKey := range abKeyed {
				if abKey != nil {
					if abText := abKey.GetLocalizedString(locKey); abText != "" {
						data.Abilities[idx][locKey] = abText
					}
				}
			}

			// SensorText
			if sensorKeyed != nil {
				if sensorText := sensorKeyed.GetLocalizedString(locKey); sensorText != "" {
					if data.SensorText == nil {
						data.SensorText = make(map[string]string)
					}
					data.SensorText[locKey] = sensorText
				}
			}

			// SimplifiedSensorText
			if simplifiedSensorKeyed != nil {
				if simplifiedSensorText := simplifiedSensorKeyed.GetLocalizedString(locKey); simplifiedSensorText != "" {
					if data.SimplifiedSensorText == nil {
						data.SimplifiedSensorText = make(map[string]string)
					}
					data.SimplifiedSensorText[locKey] = simplifiedSensorText
				}
			}

			// ScanText
			if scanKeyed != nil {
				if scanText := scanKeyed.GetLocalizedString(locKey); scanText != "" {
					if data.ScanText == nil {
						data.ScanText = make(map[string]string)
					}
					data.ScanText[locKey] = scanText
				}
			}

			// SimplifiedScanText
			if simplifiedScanKeyed != nil {
				if simplifiedScanText := simplifiedScanKeyed.GetLocalizedString(locKey); simplifiedScanText != "" {
					if data.SimplifiedScanText == nil {
						data.SimplifiedScanText = make(map[string]string)
					}
					data.SimplifiedScanText[locKey] = simplifiedScanText
				}
			}

			// Weapons
			if _, isWeapons := obj.(*WeaponsNameTextObject); isWeapons {
				if data.Weapons == nil {
					data.Weapons = make(map[string]WeaponTexts, len(weaponRefs))
				}
				for _, ref := range weaponRefs {
					if _, exists := data.Weapons[ref.name]; !exists {
						data.Weapons[ref.name] = WeaponTexts{
							Name:           make(map[string]string),
							SimplifiedName: make(map[string]string),
						}
					}
					entry := data.Weapons[ref.name]
					if seg := obj.GetKeyedString(ref.key); seg != nil {
						if text := seg.GetLocalizedString(locKey); text != "" {
							entry.Name[locKey] = text
						}
					}
					if seg := obj.GetKeyedString("s" + ref.key); seg != nil {
						if text := seg.GetLocalizedString(locKey); text != "" {
							entry.SimplifiedName[locKey] = text
						}
					}
					data.Weapons[ref.name] = entry
				}
			}
		}

		hasData := len(data.Name) > 0 || len(data.SimplifiedName) > 0 ||
			len(data.Description) > 0 || len(data.SimplifiedDescription) > 0 || len(data.Effect) > 0
		if !hasData {
			for _, ab := range data.Abilities {
				if len(ab) > 0 {
					hasData = true
					break
				}
			}
		}
		if !hasData {
			hasData = len(data.SensorText) > 0 || len(data.SimplifiedSensorText) > 0 ||
				len(data.ScanText) > 0 || len(data.SimplifiedScanText) > 0
		}
		if !hasData {
			hasData = len(data.Weapons) > 0
		}

		if hasData {
			jsonEntries = append(jsonEntries, data)
		}
	})

	return createJSON(jsonEntries, jsonFileName)
}

// serializeNameOnlyToJSON converts IList data to JSON format for name-only objects.
func serializeNameOnlyToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var nameOnlyData []*JSONEntry

	objects.RangeIndex(func(i int, nameObj datastore.IGlobalLocalizedTextObject) {
		if nameObj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &JSONEntry{
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

	return createJSON(nameOnlyData, jsonFileName)
}

// ExportKeyItemsToJSON exports key items data from objectsfile.KEY_ITEMS to a JSON file.
func ExportKeyItemsToJSON() error {
	keyItems := datastore.KeyItems
	if keyItems == nil || keyItems.IsEmpty() {
		return fmt.Errorf("KEY_ITEMS data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(keyItems, "key_items_all_localizations.json")
}

// ExportCommandsToJSON exports commands data from objectsfile.COMMANDS to a JSON file.
func ExportCommandsToJSON() error {
	if datastore.Commands.IsEmpty() {
		return fmt.Errorf("COMMANDS data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(datastore.Commands, "commands_all_localizations.json")
}

// ExportItemsToJSON exports items data from objectsfile.ITEMS to a JSON file.
func ExportItemsToJSON() error {
	if datastore.Items.IsEmpty() {
		return fmt.Errorf("ITEMS data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(datastore.Items, "items_all_localizations.json")
}

// ExportArmsToJSON exports arms data from objectsfile.ARMS_TEXT to a JSON file.
func ExportArmsToJSON() error {
	if datastore.ArmsTxt.IsEmpty() {
		return fmt.Errorf("ARMS_TEXT data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(datastore.ArmsTxt, "arms_all_localizations.json")
}

// ExportConfigToJSON exports config data from objectsfile.CONFIG_TEXT to a JSON file.
func ExportConfigToJSON() error {
	if CONFIG_TEXT == nil || CONFIG_TEXT.IsEmpty() {
		return fmt.Errorf("CONFIG_TEXT data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(CONFIG_TEXT, "config_text_all_localizations.json")
}

// ExportItemCommandsToJSON exports item descriptions from objectsfile.ITEM_TEXT to a JSON file.
func ExportItemCommandsToJSON() error {
	if ITEM_TEXT == nil || ITEM_TEXT.IsEmpty() {
		return fmt.Errorf("ITEM_TEXT data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(ITEM_TEXT, "item_commands_all_localizations.json")
}

// ExportMainMenuToJSON exports main menu data from objectsfile.MMAIN_TEXT to a JSON file.
func ExportMainMenuToJSON() error {
	if MMAIN_TEXT == nil || MMAIN_TEXT.IsEmpty() {
		return fmt.Errorf("MMAIN_TEXT data not loaded or empty")
	}

	return ExportToJSON(MMAIN_TEXT, "main_menu_all_localizations.json")
}

// ExportPlayerRoomToJSON exports player room data from objectsfile.PLAYER_ROOM to a JSON file.
func ExportPlayerRoomToJSON() error {
	if PLAYER_ROOM == nil || PLAYER_ROOM.IsEmpty() {
		return fmt.Errorf("PLAYER_ROOM data not loaded or empty")
	}

	return ExportToJSON(PLAYER_ROOM, "player_room_all_localizations.json")
}

// ExportBuildToJSON exports build data from objectsfile.BUILD_TEXT to a JSON file.
func ExportBuildToJSON() error {
	if BUILD_TEXT == nil || BUILD_TEXT.IsEmpty() {
		return fmt.Errorf("BUILD_TEXT data not loaded or empty")
	}

	return ExportToJSON(BUILD_TEXT, "build_all_localizations.json")
}

// ExportBattleToJSON exports battle data from objectsfile.BTL_TEXT to a JSON file.
func ExportBattleToJSON() error {
	if datastore.BattleTxt.IsEmpty() {
		return fmt.Errorf("BTL_TEXT data not loaded or empty")
	}

	return ExportToJSON(datastore.BattleTxt, "battle_text_all_localizations.json")
}

// ExportBattleEndToJSON exports battle end data from objectsfile.BTLEND_TEXT to a JSON file.
func ExportBattleEndToJSON() error {
	if datastore.BattleEndTxt.IsEmpty() {
		return fmt.Errorf("BTLEND_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(datastore.BattleEndTxt, "battle_end_text_all_localizations.json")
}

// ExportMonsterMagic1ToJSON exports monster magic 1 data from objectsfile.MONMAGIC1 to a JSON file.
func ExportMonsterMagic1ToJSON() error {
	if MONMAGIC1 == nil || MONMAGIC1.IsEmpty() {
		return fmt.Errorf("MONMAGIC1 data not loaded or empty")
	}

	return serializeNameOnlyToJSON(MONMAGIC1, "monster_magic1_all_localizations.json")
}

// getLocalizationKeys returns all available localization keys
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

// buildEventStringDataJSON creates EventStringDataJSON from an event string with all localizations
func buildEventStringDataJSON(index int, str interface{ GetLocalizedString(string) string }, localizationKeys []string) EventStringDataJSON {
	stringData := EventStringDataJSON{
		Index: index,
		Text:  make(map[string]string),
	}

	for _, langKey := range localizationKeys {
		value := str.GetLocalizedString(langKey)
		stringData.Text[langKey] = value
	}

	return stringData
}

// processEventFromMemory processes an event from memory and creates EventFileDataJSON
func processEventFromMemory(eventID string, localizationKeys []string) *EventFileDataJSON {
	eventFile := event.GetEvent(eventID)
	if eventFile == nil || eventFile.Strings == nil || len(eventFile.Strings) == 0 {
		return nil
	}

	eventData := EventFileDataJSON{
		ID:      eventFile.ID,
		Strings: make([]EventStringDataJSON, 0, len(eventFile.Strings)),
	}

	for i, str := range eventFile.Strings {
		stringData := buildEventStringDataJSON(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	if len(eventData.Strings) == 0 {
		common.LogVerbose("No strings found for event %s, skipping", eventID)
		return nil
	}

	return &eventData
}

// processEventFromFile processes an event from file and creates EventFileDataJSON
func processEventFromFile(eventID string, localizationKeys []string) *EventFileDataJSON {
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

	eventData := EventFileDataJSON{
		ID:      eventID,
		Strings: make([]EventStringDataJSON, 0, len(eventFileStrings)),
	}

	for i, str := range eventFileStrings {
		stringData := buildEventStringDataJSON(i, str, localizationKeys)
		eventData.Strings = append(eventData.Strings, stringData)
	}

	return &eventData
}

// writeEventJSONFile writes event data to a JSON file
func writeEventJSONFile(events []EventFileDataJSON, fileName string) error {
	if len(events) == 0 {
		common.LogVerbose("No events with string data found to export to JSON")
		return nil
	}

	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	filePath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	export := make([]models.EventFileExport, 0, len(events))
	for _, e := range events {
		strings := make([]models.EventStringDataExport, 0, len(e.Strings))
		for _, s := range e.Strings {
			strings = append(strings, models.EventStringDataExport{Index: s.Index, Text: s.Text})
		}
		export = append(export, models.EventFileExport{
			Metadata: models.NewFileMetadata(models.NewFileInfoFromPath(models.EventBinaryPath(e.ID))),
			ID:       e.ID,
			Strings:  strings,
		})
	}

	if err := models.SaveDataFile(export, filePath); err != nil {
		return fmt.Errorf("error writing JSON file %s: %w", filePath, err)
	}

	common.LogVerbose("Exported event JSON file: %s", filePath)
	common.LogVerbose("Total events exported: %d", len(events))
	return nil
}

// ExportAllEventsToJSON exports all event data to a JSON file with all localizations.
func ExportAllEventsToJSON() error {
	fileName := "events_all_localizations.json"
	localizationKeys := getSortedLocalizationKeys()
	eventIDs := event.GetAllEventIDs()

	var allEvents []EventFileDataJSON
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
func ExportEventsForLocalizationToJSON(languageCode string) error {
	eventIDs := event.GetAllEventIDs()
	localizationKeys := []string{languageCode}

	var allEvents []EventFileDataJSON

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
func ExportSingleEventToJSON(eventId string) error {
	localizationKeys := getSortedLocalizationKeys()

	eventData := processEventFromFile(eventId, localizationKeys)
	if eventData == nil {
		return fmt.Errorf("no data found for event %s", eventId)
	}

	allEvents := []EventFileDataJSON{*eventData}
	fileName := "event_" + eventId + "_all_localizations.json"

	return writeEventJSONFile(allEvents, fileName)
}
