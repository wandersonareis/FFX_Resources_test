package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/datastore"
)

// ReadNameDescriptionLocalizations reads battle command data from the command.bin file
// and loads all available localizations for each command entry directly into COMMANDS.
//
// This function reads CommandDataObject entries containing name and description information
// for combat abilities and skills. Each command includes localized text for all supported
// languages in the game. The data is loaded directly into the COMMANDS variable.
//
// File format: name and description data
// Pattern path: ex: "battle/kernel/command.bin"
func ReadNameOnlyLocalizations(patternPath string) datastore.IBinaryFile {
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameOnlyTextObjectV2(cBytes, sBytes, hLen, lang)
		}
		return NewNameOnlyTextObject(cBytes, sBytes, hLen, lang)
	}

	binaryDataFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if binaryDataFile.Objects != nil && !binaryDataFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", binaryDataFile.Objects.Len())
		datastore.Commands = binaryDataFile.GetObjects()
	}
	return binaryDataFile
}

// ReadDescriptionOnlyLocalizations reads description-only data from a binary file
// and loads all available localizations for each entry.
//
// This function reads DescriptionOnlyTextObject entries containing description and
// simplified description information. Each entry includes localized text for all
// supported languages in the game.
//
// Only applicable for FFX (v1) game version. Returns nil for other versions.
//
// File format: description only data (binary format)
// Pattern path: Variable, passed as parameter
func ReadDescriptionOnlyLocalizations(patternPath string) datastore.IBinaryFile {
	if common.GetGameVersionString() != "ffx" {
		common.LogVerbose("ReadDescriptionOnlyLocalizations is only applicable for FFX (v1) game version.")
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		return NewDescriptionOnlyTextObject(cBytes, sBytes, hLen, lang)
	}

	binaryDataFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading description-only binary data: %v", err)
		return nil
	}

	if binaryDataFile.Objects != nil && !binaryDataFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d description-only objects with all localizations", binaryDataFile.Objects.Len())
	}

	return binaryDataFile
}

// ReadNameDescriptionLocalizations reads battle command data from the command.bin file
// and loads all available localizations for each command entry directly into COMMANDS.
//
// This function reads CommandDataObject entries containing name and description information
// for combat abilities and skills. Each command includes localized text for all supported
// languages in the game. The data is loaded directly into the COMMANDS variable.
//
// File format: name and description data
// Pattern path: ex: "battle/kernel/command.bin"
func ReadNameDescriptionLocalizations(patternPath string) datastore.IBinaryFile {
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameDescriptionTextObjectV2(cBytes, sBytes, hLen, lang)
		}
		return NewNameDescriptionTextObject(cBytes, sBytes, hLen, lang)
	}

	binaryDataFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if binaryDataFile.Objects != nil && !binaryDataFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", binaryDataFile.Objects.Len())
		datastore.Commands = binaryDataFile.GetObjects()
	}
	return binaryDataFile
}

func ReadNameDescriptionEffectLocalizations(patternPath string, effectSegmentPosition int64) datastore.IBinaryFile {
	if common.GetGameVersionString() != "ffx2" {
		common.LogVerbose("ReadNameDescriptionEffectLocalizations is only applicable for FFX-2 (v2) game version.")
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameDescriptionEffect(cBytes, sBytes, hLen, effectSegmentPosition, lang)
		}
		return NewNameDescriptionTextObject(cBytes, sBytes, hLen, lang)
	}

	binaryDataFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if binaryDataFile.Objects != nil && !binaryDataFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", binaryDataFile.Objects.Len())
		datastore.Commands = binaryDataFile.GetObjects()
	}
	return binaryDataFile
}

func ReadNameDescriptionEffectAbilitiesLocalizations(patternPath string, abilitiesCount int,
	effectSegmentPosition int64) datastore.IBinaryFile {
	if common.GetGameVersionString() != "ffx2" {
		common.LogVerbose("ReadNameDescriptionEffectLocalizations is only applicable for FFX-2 (v2) game version.")
		return nil
	}

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameDescriptionEffectAbility(cBytes, sBytes, hLen, abilitiesCount, effectSegmentPosition, lang)
		}
		return nil, nil
	}

	binaryDataFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if binaryDataFile.Objects != nil && !binaryDataFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", binaryDataFile.Objects.Len())
		datastore.Commands = binaryDataFile.GetObjects()
	}
	return binaryDataFile
}

func ReadNameSensorScanLocalizations(patternPath string) datastore.IBinaryFile {
	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx" {
			return NewNameSensorScan(cBytes, sBytes, hLen, lang)
		}
		return nil, nil
	}

	binaryDataFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := binaryDataFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading commands binary data: %v", err)
		return nil
	}

	if binaryDataFile.Objects != nil && !binaryDataFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d commands with all localizations", binaryDataFile.Objects.Len())
		datastore.Commands = binaryDataFile.GetObjects()
	}
	return binaryDataFile
}

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
		patternPath,
		func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
			return NewNameDescriptionTextObjectV2(cBytes, sBytes, hLen, lang)
		},
		common.DefaultLocalization,
	)

	if err := keyItemsFile.LoadFromBinary(); err != nil {
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
func ReadItemsWithAllLocalizations() datastore.IBinaryFile {
	patternPath := "battle/kernel/item.bin"

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameDescriptionTextObjectV2(cBytes, sBytes, hLen, lang)
		}
		return NewNameDescriptionTextObject(cBytes, sBytes, hLen, lang)
	}

	itemsFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := itemsFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading items binary data: %v", err)
		return nil
	}

	if itemsFile.Objects != nil && !itemsFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d items with all localizations", itemsFile.Objects.Len())
		datastore.Items = itemsFile.GetObjects()
	}
	return itemsFile
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
func ReadPlayerRomTextWithAllLocalizations() datastore.IBinaryFile {
	patternPath := "battle/kernel/ply_rom.bin"

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameOnlyTextObjectV2(cBytes, sBytes, hLen, lang)
		}
		return NewNameOnlyTextObject(cBytes, sBytes, hLen, lang)
	}

	playerRoomTextFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := playerRoomTextFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading player room text binary data: %v", err)
		return nil
	}

	if playerRoomTextFile.Objects != nil && !playerRoomTextFile.Objects.IsEmpty() {
		common.LogVerbose("Loaded %d player room text entries with all localizations", playerRoomTextFile.Objects.Len())
	}
	return playerRoomTextFile
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
func ReadBattleTextWithAllLocalizations() datastore.IBinaryFile {
	patternPath := "battle/kernel/btl_txt.bin"

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameOnlyTextObjectV2(cBytes, sBytes, hLen, lang)
		}
		return NewNameOnlyTextObject(cBytes, sBytes, hLen, lang)
	}

	battleTextFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := battleTextFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading oversoul binary data: %v", err)
		return nil
	}

	if battleTextFile.Objects != nil && !battleTextFile.Objects.IsEmpty() {
		datastore.BattleTxt = battleTextFile.GetObjects()
		common.LogVerbose("Loaded %d battle text entries with all localizations", battleTextFile.Objects.Len())
	}
	return battleTextFile
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
func ReadBattleEndTextWithAllLocalizations() datastore.IBinaryFile {
	patternPath := "battle/kernel/btlend_txt.bin"

	creatorFunc := func(cBytes, sBytes []byte, hLen int, lang string) (datastore.IGlobalLocalizedTextObject, error) {
		if common.GetGameVersionString() == "ffx2" {
			return NewNameOnlyTextObjectV2(cBytes, sBytes, hLen, lang)
		}
		return NewNameOnlyTextObject(cBytes, sBytes, hLen, lang)
	}

	battleEndTextFile := NewBinaryFile(
		patternPath,
		creatorFunc,
		common.DefaultLocalization,
	)

	if err := battleEndTextFile.LoadFromBinary(); err != nil {
		common.LogVerbose("Error loading battle end text binary data: %v", err)
		return nil
	}

	if battleEndTextFile.Objects != nil && !battleEndTextFile.Objects.IsEmpty() {
		datastore.BattleEndTxt = battleEndTextFile.GetObjects()
		common.LogVerbose("Loaded %d battle end text entries with all localizations", battleEndTextFile.Objects.Len())
	}
	return battleEndTextFile
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
	MONMAGIC1 = ReadNameOnlyTextObjectsWithIlist(patternPath)

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
	MONMAGIC2 = ReadNameOnlyTextObjectsWithIlist(patternPath)

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
	BUILD_TEXT = ReadNameOnlyTextObjectsWithIlist(patternPath)

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
	NAME_TEXT = ReadNameOnlyTextObjectsWithIlist(patternPath)

	if NAME_TEXT != nil {
		common.LogVerbose("Loaded %d name text entries with all localizations", NAME_TEXT.Len())
	}
}
