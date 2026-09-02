package objectsfile

import (
	"path/filepath"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/datastore"
)

// ReadCommandsWithAllLocalizations reads battle command data from the command.bin file
// and loads all available localizations for each command entry directly into COMMANDS.
//
// This function reads CommandDataObject entries containing name and description information
// for combat abilities and skills. Each command includes localized text for all supported
// languages in the game. The data is loaded directly into the COMMANDS variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/command.bin"
func ReadCommandsWithAllLocalizations() {
	patternPath := "battle/kernel/command.bin"
	datastore.Commands = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d commands with all localizations", datastore.Commands.Len())
}

// ReadKeyItemsWithAllLocalizations reads key item data from the important.bin file
// and loads all available localizations for each key item entry directly into the datastore.
//
// This function reads LocalizedTextObject entries containing name and description information
// for important story items and quest items. Each key item includes localized text for all supported
// languages in the game. The data is loaded directly into the global datastore via the KEY_ITEMS adapter.
//
// File format: name and description data
// Pattern path: "battle/kernel/important.bin"
//
// Deprecated: Use the new version that returns datastore.IBinaryFile for better lifecycle management.
/* func ReadKeyItemsWithAllLocalizations() {
	patternPath := "battle/kernel/important.bin"
	keyItems := ReadNameDescriptionObjectsWithIlist(patternPath)

	if keyItems != nil {
		common.LogVerbose("Loaded %d key items with all localizations", keyItems.Len())
		datastore.KeyItems = keyItems
	}
} */

// ReadKeyItemsWithAllLocalizations reads key item data from the important.bin file
// using the BinaryFile orchestrator for lifecycle management.
//
// This function creates a BinaryFile with the appropriate creator function for
// NameDescriptionTextObjectV2 chunks, reads the binary data, and populates the
// global datastore.KeyItems with the parsed objects.
//
// File format: name and description data (V2 format)
// Pattern path: "battle/kernel/important.bin"
func ReadKeyItemsWithAllLocalizations() datastore.IBinaryFile {
	patternPath := "battle/kernel/important.bin"

	keyItemsFile := NewBinaryFile(
		func(cBytes, sBytes []byte, hLen int, lang string) datastore.IGlobalLocalizedTextObject {
			return NewNameDescriptionTextObjectV2(cBytes, sBytes, hLen, lang)
		},
		nil,
		func(fileName string, objectsList components.IList[datastore.IGlobalLocalizedTextObject]) error {
			return ImportLocalizedDataFromJsonFile(fileName, objectsList)
		},
	)

	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)
	binaryData, err := common.ReadFile(filePath)
	if err != nil {
		common.LogVerbose("Error reading key items binary file: %v", err)
		return keyItemsFile
	}

	if err := keyItemsFile.LoadFromBinary(binaryData); err != nil {
		common.LogVerbose("Error loading key items binary data: %v", err)
		return keyItemsFile
	}

	if keyItemsFile.Objects != nil && !keyItemsFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d key items with all localizations", keyItemsFile.Objects.Len())
		datastore.KeyItems = keyItemsFile.GetObjects()
	}

	return keyItemsFile
}

// ReadItemsWithAllLocalizations reads item data from the item.bin file
// and loads all available localizations for each item entry directly into ITEMS.
//
// This function reads LocalizedTextObject entries containing name and description information
// for consumable items, equipment, and other usable items. Each item includes localized text for all supported
// languages in the game. The data is loaded directly into the global ITEMS variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/item.bin"
func ReadItemsWithAllLocalizations() {
	patternPath := "battle/kernel/item.bin"
	datastore.Items = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d items with all localizations", datastore.Items.Len())
}

// ReadArmsTextWithAllLocalizations reads arms text data from the arms_txt.bin file
// and loads all available localizations for each arms text entry directly into ARMS_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for weapon and armor names and descriptions. Each arms text includes localized text for all supported
// languages in the game. The data is loaded directly into the global ARMS_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/arms_txt.bin"
func ReadArmsTextWithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/arms_txt.bin"
	datastore.ArmsTxt = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d arms text entries with all localizations", datastore.ArmsTxt.Len())
}

// ReadConfigTextWithAllLocalizations reads config text data from the config_txt.bin file
// and loads all available localizations for each config text entry directly into CONFIG_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for configuration settings and system messages. Each config text includes localized text for all supported
// languages in the game. The data is loaded directly into the global CONFIG_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/config_txt.bin"
func ReadConfigTextWithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/config_txt.bin"
	CONFIG_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	if CONFIG_TEXT != nil {
		common.LogVerbose("Loaded %d config text entries with all localizations", CONFIG_TEXT.Len())
	}
}

// ReadItemCommandsWithAllLocalizations reads item text data from the item_txt.bin file
// and loads all available localizations for each item text entry directly into ITEM_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for item text and detailed descriptions. Each item text includes localized text for all supported
// languages in the game. The data is loaded directly into the global ITEM_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/item_txt.bin"
func ReadItemCommandsWithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/item_txt.bin"
	ITEM_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	if ITEM_TEXT != nil {
		common.LogVerbose("Loaded %d item commands entries with all localizations", ITEM_TEXT.Len())
	}
}

// ReadMainMenuTextWithAllLocalizations reads main menu text data from the mmain_txt.bin file
// and loads all available localizations for each main menu text entry directly into MMAIN_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for main menu text and interface messages. Each main menu text includes localized text for all supported
// languages in the game. The data is loaded directly into the global MMAIN_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/mmain_txt.bin"
func ReadMainMenuTextWithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/mmain_txt.bin"
	MMAIN_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	if MMAIN_TEXT != nil {
		common.LogVerbose("Loaded %d main menu text entries with all localizations", MMAIN_TEXT.Len())
	}
}

// ReadPlayerRomTextWithAllLocalizations reads player ROM text data from the ply_rom.bin file
// and loads all available localizations for each player ROM text entry directly into PLAYER_ROOM.
//
// This function reads LocalizedTextObject entries containing name and description information
// for player ROM data and character-related text. Each player ROM text includes localized text for all supported
// languages in the game. The data is loaded directly into the global PLAYER_ROOM variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/ply_rom.bin"
func ReadPlayerRomTextWithAllLocalizations() {
	patternPath := "battle/kernel/ply_rom.bin"
	PLAYER_ROOM = ReadNameOnlyDataObjectsWithIlist(patternPath)

	if PLAYER_ROOM != nil {
		common.LogVerbose("Loaded %d player ROM text entries with all localizations", PLAYER_ROOM.Len())
	}
}

// ReadPlayerSaveWithAllLocalizations reads player save data from the ply_save.bin file (FFX-2 only).
func ReadPlayerSaveWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/ply_save.bin"
	PLAYER_SAVE = ReadNameDescriptionObjectsWithIlist(patternPath)

	if PLAYER_SAVE != nil {
		common.LogVerbose("Loaded %d player save entries with all localizations", PLAYER_SAVE.Len())
	}
}

// ReadBattleTextWithAllLocalizations reads battle text data from the btl_txt.bin file
// and loads all available localizations for each battle text entry directly into BTL_TEXT.
//
// This function reads LocalizedTextObject entries containing name information
// for battle-related text and messages. Each battle text includes localized text for all supported
// languages in the game. The data is loaded directly into the global BTL_TEXT variable.
//
// File format: name only data
// Pattern path: "battle/kernel/btl_txt.bin"
func ReadBattleTextWithAllLocalizations() {
	patternPath := "battle/kernel/btl_txt.bin"
	if common.GetGameVersionString() == "ffx2" {
		datastore.BattleTxt = ReadNameOnlyDataObjectsWithIlist(patternPath)
	} else {
		datastore.BattleTxt = ReadNameOnlyDataObjectsWithIlist(patternPath)
	}

	common.LogVerbose("Loaded %d battle text entries with all localizations", datastore.BattleTxt.Len())
}

// ReadBattleEndTextWithAllLocalizations reads battle end text data from the btlend_txt.bin file
// and loads all available localizations for each battle end text entry directly into BTLEND_TEXT.
//
// This function reads LocalizedTextObject entries containing name information
// for battle end text and messages. Each battle end text includes localized text for all supported
// languages in the game. The data is loaded directly into the global BTLEND_TEXT variable.
//
// File format: name only data
// Pattern path: "battle/kernel/btlend_txt.bin"
func ReadBattleEndTextWithAllLocalizations() {
	patternPath := "battle/kernel/btlend_txt.bin"
	datastore.BattleEndTxt = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d battle end text entries with all localizations", datastore.BattleEndTxt.Len())
}

// ReadMonsterMagic1WithAllLocalizations reads monster magic 1 data from the monmagic1.bin file
// and loads all available localizations for each monster magic entry directly into MONMAGIC1.
//
// This function reads LocalizedTextObject entries containing name information
// for monster magic spells and abilities. Each monster magic includes localized text for all supported
// languages in the game. The data is loaded directly into the global MONMAGIC1 variable.
//
// File format: name only data
// Pattern path: "battle/kernel/monmagic1.bin"
func ReadMonsterMagic1WithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/monmagic1.bin"
	MONMAGIC1 = ReadNameOnlyDataObjectsWithIlist(patternPath)

	if MONMAGIC1 != nil {
		common.LogVerbose("Loaded %d monster magic 1 entries with all localizations", MONMAGIC1.Len())
	}
}

// ReadMonsterMagic2WithAllLocalizations reads monster magic 2 data from the monmagic2.bin file
// and loads all available localizations for each monster magic entry directly into MONMAGIC2.
//
// This function reads LocalizedTextObject entries containing name information
// for monster magic spells and abilities. Each monster magic includes localized text for all supported
// languages in the game. The data is loaded directly into the global MONMAGIC2 variable.
//
// File format: name only data
// Pattern path: "battle/kernel/monmagic2.bin"
func ReadMonsterMagic2WithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/monmagic2.bin"
	MONMAGIC2 = ReadNameOnlyDataObjectsWithIlist(patternPath)

	if MONMAGIC2 != nil {
		common.LogVerbose("Loaded %d monster magic 2 entries with all localizations", MONMAGIC2.Len())
	}
}

// ReadBuildTextWithAllLocalizations reads build text data from the build_txt.bin file
// and loads all available localizations for each build text entry directly into BUILD_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for build-related text and construction messages. Each build text includes localized text for all supported
// languages in the game. The data is loaded directly into the global BUILD_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/build_txt.bin"
func ReadBuildTextWithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/build_txt.bin"
	BUILD_TEXT = ReadNameOnlyDataObjectsWithIlist(patternPath)

	if BUILD_TEXT != nil {
		common.LogVerbose("Loaded %d build text entries with all localizations", BUILD_TEXT.Len())
	}
}

// ReadNameTextWithAllLocalizations reads name text data from the name_txt.bin file
// and loads all available localizations for each name text entry directly into NAME_TEXT.
//
// This function reads LocalizedTextObject entries containing name information
// for name text and character names. Each name text includes localized text for all supported
// languages in the game. The data is loaded directly into the global NAME_TEXT variable.
//
// File format: name only data
// Pattern path: "battle/kernel/name_txt.bin"
func ReadNameTextWithAllLocalizations() {
	if common.GetGameVersionString() == "ffx2" {
		return
	}
	patternPath := "battle/kernel/name_txt.bin"
	NAME_TEXT = ReadNameOnlyDataObjectsWithIlist(patternPath)

	if NAME_TEXT != nil {
		common.LogVerbose("Loaded %d name text entries with all localizations", NAME_TEXT.Len())
	}
}

// As funções abaixo são exclusivas do FFX-2 (v2). Retornam cedo se a versão do
// jogo não for FFX-2. Todas usam ReadNameDescriptionObjectsWithIlist, que é a
// única leitora atualizada para o formato V2 dos arquivos de objetos.

// ReadAAbilityWithAllLocalizations reads ability data from the a_ability.bin file (FFX-2 only).
func ReadAAbilityWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/a_ability.bin"
	A_ABILITY = ReadNameDescriptionObjectsWithIlist(patternPath)

	if A_ABILITY != nil {
		common.LogVerbose("Loaded %d a-ability entries with all localizations", A_ABILITY.Len())
	}
}

// ReadAccessoriesWithAllLocalizations reads accessory data from the accessory.bin file (FFX-2 only).
func ReadAccessoriesWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/accessory.bin"
	ACCESSORY = ReadNameDescriptionObjectsWithIlist(patternPath)

	if ACCESSORY != nil {
		common.LogVerbose("Loaded %d accessory entries with all localizations", ACCESSORY.Len())
	}
}

// ReadJobsWithAllLocalizations reads job data from the job.bin file (FFX-2 only).
func ReadJobsWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/job.bin"
	JOB = ReadJobObjectsWithIlist(patternPath)

	if JOB != nil {
		common.LogVerbose("Loaded %d job entries with all localizations", JOB.Len())
	}
}

// ReadMenuTextWithAllLocalizations reads menu text data from the menu_txt.bin file (FFX-2 only).
func ReadMenuTextWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/menu_txt.bin"
	MENU_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	if MENU_TEXT != nil {
		common.LogVerbose("Loaded %d menu text entries with all localizations", MENU_TEXT.Len())
	}
}

// ReadMonsterMagicWithAllLocalizations reads monster magic data from the monmagic.bin file (FFX-2 only).
func ReadMonsterMagicWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/monmagic.bin"
	MONMAGIC = ReadNameDescriptionObjectsWithIlist(patternPath)

	if MONMAGIC != nil {
		common.LogVerbose("Loaded %d monster magic entries with all localizations", MONMAGIC.Len())
	}
}

// ReadMonstersWithAllLocalizations reads monster data from the monster.bin file
// and loads all available localizations for each monster entry directly into MONSTER.
func ReadMonstersWithAllLocalizations() {
	patternPath := "battle/kernel/monster.bin"
	MONSTER = ReadNameDescriptionObjectsWithIlist(patternPath)

	if MONSTER != nil {
		common.LogVerbose("Loaded %d monster entries with all localizations", MONSTER.Len())
	}
}

// ReadMonsters2WithAllLocalizations reads monster data from the monster2.bin file (FFX-2 only).
func ReadMonsters2WithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/monster2.bin"
	MONSTER2 = ReadNameDescriptionObjectsWithIlist(patternPath)

	if MONSTER2 != nil {
		common.LogVerbose("Loaded %d monster2 entries with all localizations", MONSTER2.Len())
	}
}

// ReadOversoulWithAllLocalizations reads oversoul data from the oversoul.bin file (FFX-2 only).
func ReadOversoulWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/oversoul.bin"
	OVERSOUL = ReadNameOnlyDataObjectsWithIlist(patternPath)

	if OVERSOUL != nil {
		common.LogVerbose("Loaded %d oversoul entries with all localizations", OVERSOUL.Len())
	}
}

// ReadPlateWithAllLocalizations reads plate data from the plate.bin file (FFX-2 only).
func ReadPlateWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/plate.bin"
	PLATE = ReadPlateObjectsWithIlist(patternPath)

	if PLATE != nil {
		common.LogVerbose("Loaded %d plate entries with all localizations", PLATE.Len())
	}
}

// ReadPlateObjectsWithIlist reads plate binary data and creates PlateTextObject entries
// with all available localizations. The plate format includes name, description, 4 abilities,
// and effect field per entry.
func ReadPlateObjectsWithIlist(patternPath string) components.IList[datastore.IGlobalLocalizedTextObject] {
	filePath := filepath.Join(common.GetLocalizationRoot(common.DefaultLocalization), patternPath)

	creator := func(data []byte, stringBytes []byte, headerLength int, localization string) datastore.IGlobalLocalizedTextObject {
		if obj := NewPlateTextObject(data, stringBytes, headerLength, localization); obj != nil {
			return obj
		}
		return nil
	}

	var plateObjects components.IList[datastore.IGlobalLocalizedTextObject]
	if common.GetGameVersionString() == "ffx2" {
		plateObjects = ReadDataListWithIlistV2(filePath, common.DefaultLocalization, creator)
	} else {
		plateObjects = ReadDataListWithIlist(filePath, common.DefaultLocalization, creator)
	}
	if plateObjects == nil || plateObjects.IsEmpty() {
		common.LogVerbose("No plate objects found for %s\n", patternPath)
		return components.NewList[datastore.IGlobalLocalizedTextObject](0)
	}

	PopulateDataObjectLocalizationsWithIlist(patternPath, plateObjects, creator)

	common.LogVerbose("Loading %d plate entries...\n", plateObjects.Len())

	return plateObjects
}

// ReadSaveTextWithAllLocalizations reads save text data from the save_txt.bin file (FFX-2 only).
func ReadSaveTextWithAllLocalizations() {
	if common.GetGameVersionString() != "ffx2" {
		return
	}
	patternPath := "battle/kernel/save_txt.bin"
	SAVE_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	if SAVE_TEXT != nil {
		common.LogVerbose("Loaded %d save text entries with all localizations", SAVE_TEXT.Len())
	}
}
