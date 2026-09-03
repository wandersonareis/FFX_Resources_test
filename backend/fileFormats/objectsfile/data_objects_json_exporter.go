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

// createNameDescriptionJSON creates a JSON file with name and description data.
func createNameDescriptionJSON(nameDescriptionList []*NameDescriptionData, fileName string) error {
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

// createNameOnlyJSON creates a JSON file with name-only data.
func createNameOnlyJSON(data []*NameOnlyData, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	stringsBytes, err := json.Marshal(data)
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

	common.LogVerbose("Exported name-only data to JSON: %s", jsonPath)
	return nil
}

// ExportNameDescriptionToJSON converts IList data to JSON format for name-description objects.
func ExportNameDescriptionToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
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

// serializeJobTextToJSON converts IList data to JSON format for job text objects.
func serializeJobTextToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var jobTextData []*JobTextData

	objects.RangeIndex(func(i int, jobObj datastore.IGlobalLocalizedTextObject) {
		if jobObj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &JobTextData{
			NameOnlyData: NameOnlyData{
				ID:   i,
				Name: make(map[string]string),
			},
			Description: make(map[string]string),
			Effect:      make(map[string]string),
		}

		nameKeyed := jobObj.GetKeyedString("name")
		descKeyed := jobObj.GetKeyedString("description")
		effKeyed := jobObj.GetKeyedString("effect")

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

			if effKeyed != nil {
				effText := effKeyed.GetLocalizedString(locKey)
				if effText != "" {
					data.Effect[locKey] = effText
				}
			}
		}

		if len(data.Name) > 0 || len(data.Description) > 0 || len(data.Effect) > 0 {
			jobTextData = append(jobTextData, data)
		}
	})

	return createJobTextJSON(jobTextData, jsonFileName)
}

// createJobTextJSON creates a JSON file with job text data.
func createJobTextJSON(jobTextData []*JobTextData, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	stringsBytes, err := json.Marshal(jobTextData)
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

	common.LogVerbose("Exported job text data to JSON: %s", jsonPath)
	return nil
}

// exportPlateTextToJSON converts IList data to JSON format for plate text objects.
func exportPlateTextToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var plateTextData []*PlateTextData

	objects.RangeIndex(func(i int, plateObj datastore.IGlobalLocalizedTextObject) {
		if plateObj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &PlateTextData{
			NameOnlyData: NameOnlyData{
				ID:   i,
				Name: make(map[string]string),
			},
			Description: make(map[string]string),
			Abilities:   make([]map[string]string, 4),
			Effect:      make(map[string]string),
		}

		for k := range data.Abilities {
			data.Abilities[k] = make(map[string]string)
		}

		nameKeyed := plateObj.GetKeyedString("name")
		descKeyed := plateObj.GetKeyedString("description")
		effKeyed := plateObj.GetKeyedString("effect")

		abKeyed := make([]datastore.IGlobalLocalizedKeyedStringObject, 4)
		abKeyed[0] = plateObj.GetKeyedString("ability1")
		abKeyed[1] = plateObj.GetKeyedString("ability2")
		abKeyed[2] = plateObj.GetKeyedString("ability3")
		abKeyed[3] = plateObj.GetKeyedString("ability4")

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

			if effKeyed != nil {
				effText := effKeyed.GetLocalizedString(locKey)
				if effText != "" {
					data.Effect[locKey] = effText
				}
			}

			for idx, abKey := range abKeyed {
				if abKey != nil {
					abText := abKey.GetLocalizedString(locKey)
					if abText != "" {
						data.Abilities[idx][locKey] = abText
					}
				}
			}
		}

		hasData := len(data.Name) > 0 || len(data.Description) > 0 || len(data.Effect) > 0
		if !hasData {
			for _, ab := range data.Abilities {
				if len(ab) > 0 {
					hasData = true
					break
				}
			}
		}

		if hasData {
			plateTextData = append(plateTextData, data)
		}
	})

	return createPlateTextJSON(plateTextData, jsonFileName)
}

// createPlateTextJSON creates a JSON file with plate text data.
func createPlateTextJSON(plateTextData []*PlateTextData, fileName string) error {
	editsPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if err := common.EnsurePathExists(editsPath); err != nil {
		return fmt.Errorf("error creating edits directory: %w", err)
	}

	jsonPath := filepath.Join(editsPath, common.WithVersionSuffix(fileName))

	stringsBytes, err := json.Marshal(plateTextData)
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

	common.LogVerbose("Exported plate text data to JSON: %s", jsonPath)
	return nil
}

// serializeNameOnlyToJSON converts IList data to JSON format for name-only objects.
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

	return ExportNameDescriptionToJSON(MMAIN_TEXT, "main_menu_all_localizations.json")
}

// ExportPlayerRoomToJSON exports player room data from objectsfile.PLAYER_ROOM to a JSON file.
func ExportPlayerRoomToJSON() error {
	if PLAYER_ROOM == nil || PLAYER_ROOM.IsEmpty() {
		return fmt.Errorf("PLAYER_ROOM data not loaded or empty")
	}

	return serializeNameOnlyToJSON(PLAYER_ROOM, "player_room_all_localizations.json")
}

// ExportBuildToJSON exports build data from objectsfile.BUILD_TEXT to a JSON file.
func ExportBuildToJSON() error {
	if BUILD_TEXT == nil || BUILD_TEXT.IsEmpty() {
		return fmt.Errorf("BUILD_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(BUILD_TEXT, "build_all_localizations.json")
}

// ExportBattleToJSON exports battle data from objectsfile.BTL_TEXT to a JSON file.
func ExportBattleToJSON() error {
	if datastore.BattleTxt.IsEmpty() {
		return fmt.Errorf("BTL_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(datastore.BattleTxt, "battle_text_all_localizations.json")
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

// ExportMonsterMagic2ToJSON exports monster magic 2 data from objectsfile.MONMAGIC2 to a JSON file.
func ExportMonsterMagic2ToJSON() error {
	if MONMAGIC2 == nil || MONMAGIC2.IsEmpty() {
		return fmt.Errorf("MONMAGIC2 data not loaded or empty")
	}

	return serializeNameOnlyToJSON(MONMAGIC2, "monster_magic2_all_localizations.json")
}

// ExportNameToJSON exports name data from objectsfile.NAME_TEXT to a JSON file.
func ExportNameToJSON() error {
	if NAME_TEXT == nil || NAME_TEXT.IsEmpty() {
		return fmt.Errorf("NAME_TEXT data not loaded or empty")
	}

	return serializeNameOnlyToJSON(NAME_TEXT, "names_all_localizations.json")
}

// ExportAAbilityToJSON exports FFX-2 ability data (a_ability.bin) to a JSON file.
func ExportAAbilityToJSON() error {
	if A_ABILITY == nil || A_ABILITY.IsEmpty() {
		return fmt.Errorf("A_ABILITY data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(A_ABILITY, "a_ability_all_localizations.json")
}

// ExportAccessoriesToJSON exports FFX-2 accessory data (accessory.bin) to a JSON file.
func ExportAccessoriesToJSON() error {
	if ACCESSORY == nil || ACCESSORY.IsEmpty() {
		return fmt.Errorf("ACCESSORY data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(ACCESSORY, "accessory_all_localizations.json")
}

// ExportJobsToJSON exports FFX-2 job data (job.bin) to a JSON file.
func ExportJobsToJSON() error {
	if JOB == nil || JOB.IsEmpty() {
		return fmt.Errorf("JOB data not loaded or empty")
	}

	return serializeJobTextToJSON(JOB, "job_all_localizations.json")
}

// ExportMenuTextToJSON exports FFX-2 menu text data (menu_txt.bin) to a JSON file.
func ExportMenuTextToJSON() error {
	if MENU_TEXT == nil || MENU_TEXT.IsEmpty() {
		return fmt.Errorf("MENU_TEXT data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(MENU_TEXT, "menu_text_all_localizations.json")
}

// ExportMonsterMagicToJSON exports FFX-2 monster magic data (monmagic.bin) to a JSON file.
func ExportMonsterMagicToJSON() error {
	if MONMAGIC == nil || MONMAGIC.IsEmpty() {
		return fmt.Errorf("MONMAGIC data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(MONMAGIC, "monster_magic_all_localizations.json")
}

// ExportMonstersToJSON exports FFX-2 monster data (monster.bin) to a JSON file.
func ExportMonstersToJSON() error {
	if MONSTER == nil || MONSTER.IsEmpty() {
		return fmt.Errorf("MONSTER data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(MONSTER, "monster_all_localizations.json")
}

// ExportMonsters2ToJSON exports FFX-2 monster data (monster2.bin) to a JSON file.
func ExportMonsters2ToJSON() error {
	if MONSTER2 == nil || MONSTER2.IsEmpty() {
		return fmt.Errorf("MONSTER2 data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(MONSTER2, "monster2_all_localizations.json")
}

// ExportBattleTextToJSON exports FFX-2 battle text data (battle_txt.bin) to a JSON file.
func ExportBattleTextToJSON() error {
	if BATTLE_TEXT == nil || BATTLE_TEXT.IsEmpty() {
		return fmt.Errorf("BATTLE_TEXT data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(BATTLE_TEXT, "battle_text_all_localizations.json")
}

// ExportOversoulToJSON exports FFX-2 oversoul data (oversoul.bin) to a JSON file.
func ExportOversoulToJSON() error {
	if OVERSOUL == nil || OVERSOUL.IsEmpty() {
		return fmt.Errorf("OVERSOUL data not loaded or empty")
	}

	return serializeNameOnlyToJSON(OVERSOUL, "oversoul_all_localizations.json")
}

// ExportPlateToJSON exports FFX-2 plate data (plate.bin) to a JSON file.
func ExportPlateToJSON() error {
	if PLATE == nil || PLATE.IsEmpty() {
		return fmt.Errorf("PLATE data not loaded or empty")
	}

	return exportPlateTextToJSON(PLATE, "plate_all_localizations.json")
}

// ExportPlayerSaveToJSON exports FFX-2 player save data (ply_save.bin) to a JSON file.
func ExportPlayerSaveToJSON() error {
	if PLAYER_SAVE == nil || PLAYER_SAVE.IsEmpty() {
		return fmt.Errorf("PLAYER_SAVE data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(PLAYER_SAVE, "player_save_all_localizations.json")
}

// ExportSaveTextToJSON exports FFX-2 save text data (save_txt.bin) to a JSON file.
func ExportSaveTextToJSON() error {
	if SAVE_TEXT == nil || SAVE_TEXT.IsEmpty() {
		return fmt.Errorf("SAVE_TEXT data not loaded or empty")
	}

	return ExportNameDescriptionToJSON(SAVE_TEXT, "save_text_all_localizations.json")
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

// ExportKeyItemsToJSONWithExporter exports key items using the ExportNameDescriptionToJSON function
// This is the function to be used as JsonExporterFunc in NewBinaryFile
func ExportKeyItemsToJSONWithExporter(objects components.IList[datastore.IGlobalLocalizedTextObject], fileName string) error {
	return ExportNameDescriptionToJSON(objects, fileName)
}
