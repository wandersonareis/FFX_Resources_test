package exporters

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/fileFormats/objectsfile"
	"ffxresources/backend/models"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
)

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

// serializeJobTextToJSON converts IList data to JSON format for job text objects.
//
// Parameters:
//   - objects: IList containing JobTextObject entries with name, description, and effect
//   - jsonFileName: Name of the output JSON file
//
// Returns: error if serialization fails
func serializeJobTextToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
	if objects == nil || objects.IsEmpty() {
		return fmt.Errorf("no objects loaded or empty")
	}

	var jobTextData []*objectsfile.JobTextData

	objects.RangeIndex(func(i int, jobObj datastore.IGlobalLocalizedTextObject) {
		if jobObj == nil {
			common.LogVerbose("Object %d is nil, skipping", i)
			return
		}

		data := &objectsfile.JobTextData{
			NameOnlyData: objectsfile.NameOnlyData{
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
func createJobTextJSON(jobTextData []*objectsfile.JobTextData, fileName string) error {
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

// Defina a struct localmente se não quiser/puder alterar a do pacote objectsfile
type LocalPlateTextData struct {
    ID          int               `json:"id"`
    Name        map[string]string `json:"name"`
    Description map[string]string `json:"description"`
    Abilities   []map[string]string `json:"abilities"`
    Effect      map[string]string `json:"effect"`
}

// serializePlateTextToJSON converts IList data to JSON format for plate text objects.
func serializePlateTextToJSON(objects components.IList[datastore.IGlobalLocalizedTextObject], jsonFileName string) error {
    if objects == nil || objects.IsEmpty() {
        return fmt.Errorf("no objects loaded or empty")
    }

    var plateTextData []*LocalPlateTextData

    objects.RangeIndex(func(i int, plateObj datastore.IGlobalLocalizedTextObject) {
        if plateObj == nil {
            common.LogVerbose("Object %d is nil, skipping", i)
            return
        }

        data := &LocalPlateTextData{
            ID:          i,
            Name:        make(map[string]string),
            Description: make(map[string]string),
            Abilities:   make([]map[string]string, 4), // Inicializa o array para 4 habilidades
            Effect:      make(map[string]string),
        }

        // Inicializa os mapas de cada habilidade
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

            // Extrai o texto de cada uma das 4 habilidades
            for idx, abKey := range abKeyed {
                if abKey != nil {
                    abText := abKey.GetLocalizedString(locKey)
                    if abText != "" {
                        data.Abilities[idx][locKey] = abText
                    }
                }
            }
        }

        // Verifica se o objeto possui algum texto antes de adicioná-lo à lista
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
func createPlateTextJSON(plateTextData []*LocalPlateTextData, fileName string) error {
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

	return serializeNameOnlyToJSON(objectsfile.PLAYER_ROOM, "player_room_all_localizations.json")
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

// ExportAAbilityToJSON exports FFX-2 ability data (a_ability.bin) to a JSON file.
func ExportAAbilityToJSON() error {
	if objectsfile.A_ABILITY == nil || objectsfile.A_ABILITY.IsEmpty() {
		return fmt.Errorf("A_ABILITY data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.A_ABILITY, "a_ability_all_localizations.json")
}

// ExportAccessoriesToJSON exports FFX-2 accessory data (accessory.bin) to a JSON file.
func ExportAccessoriesToJSON() error {
	if objectsfile.ACCESSORY == nil || objectsfile.ACCESSORY.IsEmpty() {
		return fmt.Errorf("ACCESSORY data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.ACCESSORY, "accessory_all_localizations.json")
}

// ExportMenuTextToJSON exports FFX-2 menu text data (menu_txt.bin) to a JSON file.
func ExportMenuTextToJSON() error {
	if objectsfile.MENU_TEXT == nil || objectsfile.MENU_TEXT.IsEmpty() {
		return fmt.Errorf("MENU_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.MENU_TEXT, "menu_text_all_localizations.json")
}

// ExportMonsterMagicToJSON exports FFX-2 monster magic data (monmagic.bin) to a JSON file.
func ExportMonsterMagicToJSON() error {
	if objectsfile.MONMAGIC == nil || objectsfile.MONMAGIC.IsEmpty() {
		return fmt.Errorf("MONMAGIC data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.MONMAGIC, "monster_magic_all_localizations.json")
}

// ExportMonstersToJSON exports FFX-2 monster data (monster.bin) to a JSON file.
func ExportMonstersToJSON() error {
	if objectsfile.MONSTER == nil || objectsfile.MONSTER.IsEmpty() {
		return fmt.Errorf("MONSTER data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.MONSTER, "monster_all_localizations.json")
}

// ExportMonsters2ToJSON exports FFX-2 monster data (monster2.bin) to a JSON file.
func ExportMonsters2ToJSON() error {
	if objectsfile.MONSTER2 == nil || objectsfile.MONSTER2.IsEmpty() {
		return fmt.Errorf("MONSTER2 data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.MONSTER2, "monster2_all_localizations.json")
}
// ExportBattleTextToJSON exports FFX-2 battle text data (battle_txt.bin) to a JSON file.
func ExportBattleTextToJSON() error {
	if objectsfile.BATTLE_TEXT == nil || objectsfile.BATTLE_TEXT.IsEmpty() {
		return fmt.Errorf("BATTLE_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.BATTLE_TEXT, "battle_text_all_localizations.json")
}

// ExportOversoulToJSON exports FFX-2 oversoul data (oversoul.bin) to a JSON file.
func ExportOversoulToJSON() error {
	if objectsfile.OVERSOUL == nil || objectsfile.OVERSOUL.IsEmpty() {
		return fmt.Errorf("OVERSOUL data not loaded or empty")
	}

	return serializeNameOnlyToJSON(objectsfile.OVERSOUL, "oversoul_all_localizations.json")
}

// ExportPlateToJSON exports FFX-2 plate data (plate.bin) to a JSON file.
func ExportPlateToJSON() error {
	if objectsfile.PLATE == nil || objectsfile.PLATE.IsEmpty() {
		return fmt.Errorf("PLATE data not loaded or empty")
	}

	return serializePlateTextToJSON(objectsfile.PLATE, "plate_all_localizations.json")
}

// ExportPlayerSaveToJSON exports FFX-2 player save data (ply_save.bin) to a JSON file.
func ExportPlayerSaveToJSON() error {
	if objectsfile.PLAYER_SAVE == nil || objectsfile.PLAYER_SAVE.IsEmpty() {
		return fmt.Errorf("PLAYER_SAVE data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.PLAYER_SAVE, "player_save_all_localizations.json")
}

// ExportSaveTextToJSON exports FFX-2 save text data (save_txt.bin) to a JSON file.
func ExportSaveTextToJSON() error {
	if objectsfile.SAVE_TEXT == nil || objectsfile.SAVE_TEXT.IsEmpty() {
		return fmt.Errorf("SAVE_TEXT data not loaded or empty")
	}

	return serializeNameDescriptionToJSON(objectsfile.SAVE_TEXT, "save_text_all_localizations.json")
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
