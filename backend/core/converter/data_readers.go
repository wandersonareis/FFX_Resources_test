package converter

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
)

// ReadCommandsWithAllLocalizations reads battle command data from the command.bin file
// and loads all available localizations for each command entry directly into components.COMMANDS.
//
// This function reads CommandDataObject entries containing name and description information
// for combat abilities and skills. Each command includes localized text for all supported
// languages in the game. The data is loaded directly into the global COMMANDS variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/command.bin"
func ReadCommandsWithAllLocalizations() {
	patternPath := "battle/kernel/command.bin"
	components.COMMANDS = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d commands with all localizations", components.COMMANDS.GetLength())
}

// ReadKeyItemsWithAllLocalizations reads key item data from the important.bin file
// and loads all available localizations for each key item entry directly into components.KEY_ITEMS.
//
// This function reads LocalizedTextObject entries containing name and description information
// for important story items and quest items. Each key item includes localized text for all supported
// languages in the game. The data is loaded directly into the global KEY_ITEMS variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/important.bin"
func ReadKeyItemsWithAllLocalizations() {
	patternPath := "battle/kernel/important.bin"
	components.KEY_ITEMS = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d key items with all localizations", components.KEY_ITEMS.GetLength())
}

// ReadItemsWithAllLocalizations reads item data from the item.bin file
// and loads all available localizations for each item entry directly into components.ITEMS.
//
// This function reads LocalizedTextObject entries containing name and description information
// for consumable items, equipment, and other usable items. Each item includes localized text for all supported
// languages in the game. The data is loaded directly into the global ITEMS variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/item.bin"
func ReadItemsWithAllLocalizations() {
	patternPath := "battle/kernel/item.bin"
	components.ITEMS = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d items with all localizations", components.ITEMS.GetLength())
}

// ReadArmsTextWithAllLocalizations reads arms text data from the arms_txt.bin file
// and loads all available localizations for each arms text entry directly into components.ARMS_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for weapon and armor names and descriptions. Each arms text includes localized text for all supported
// languages in the game. The data is loaded directly into the global ARMS_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/arms_txt.bin"
func ReadArmsTextWithAllLocalizations() {
	patternPath := "battle/kernel/arms_txt.bin"
	components.ARMS_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d arms text entries with all localizations", components.ARMS_TEXT.GetLength())
}

// ReadConfigTextWithAllLocalizations reads config text data from the config_txt.bin file
// and loads all available localizations for each config text entry directly into components.CONFIG_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for configuration settings and system messages. Each config text includes localized text for all supported
// languages in the game. The data is loaded directly into the global CONFIG_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/config_txt.bin"
func ReadConfigTextWithAllLocalizations() {
	patternPath := "battle/kernel/config_txt.bin"
	components.CONFIG_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d config text entries with all localizations", components.CONFIG_TEXT.GetLength())
}

// ReadItemCommandsWithAllLocalizations reads item text data from the item_txt.bin file
// and loads all available localizations for each item text entry directly into components.ITEM_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for item text and detailed descriptions. Each item text includes localized text for all supported
// languages in the game. The data is loaded directly into the global ITEM_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/item_txt.bin"
func ReadItemCommandsWithAllLocalizations() {
	patternPath := "battle/kernel/item_txt.bin"
	components.ITEM_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d item commands entries with all localizations", components.ITEM_TEXT.GetLength())
}

// ReadMainMenuTextWithAllLocalizations reads main menu text data from the mmain_txt.bin file
// and loads all available localizations for each main menu text entry directly into components.MMAIN_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for main menu text and interface messages. Each main menu text includes localized text for all supported
// languages in the game. The data is loaded directly into the global MMAIN_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/mmain_txt.bin"
func ReadMainMenuTextWithAllLocalizations() {
	patternPath := "battle/kernel/mmain_txt.bin"
	components.MMAIN_TEXT = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d main menu text entries with all localizations", components.MMAIN_TEXT.GetLength())
}

// ReadPlayerRomTextWithAllLocalizations reads player ROM text data from the ply_rom.bin file
// and loads all available localizations for each player ROM text entry directly into components.PLAYER_ROOM.
//
// This function reads LocalizedTextObject entries containing name and description information
// for player ROM data and character-related text. Each player ROM text includes localized text for all supported
// languages in the game. The data is loaded directly into the global PLAYER_ROOM variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/ply_rom.bin"
func ReadPlayerRomTextWithAllLocalizations() {
	patternPath := "battle/kernel/ply_rom.bin"
	components.PLAYER_ROOM = ReadNameDescriptionObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d player ROM text entries with all localizations", components.PLAYER_ROOM.GetLength())
}

// ReadBattleTextWithAllLocalizations reads battle text data from the btl_txt.bin file
// and loads all available localizations for each battle text entry directly into components.BTL_TEXT.
//
// This function reads LocalizedTextObject entries containing name information
// for battle-related text and messages. Each battle text includes localized text for all supported
// languages in the game. The data is loaded directly into the global BTL_TEXT variable.
//
// File format: name only data
// Pattern path: "battle/kernel/btl_txt.bin"
func ReadBattleTextWithAllLocalizations() {
	patternPath := "battle/kernel/btl_txt.bin"
	components.BTL_TEXT = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d battle text entries with all localizations", components.BTL_TEXT.GetLength())
}

// ReadBattleEndTextWithAllLocalizations reads battle end text data from the btlend_txt.bin file
// and loads all available localizations for each battle end text entry directly into components.BTLEND_TEXT.
//
// This function reads LocalizedTextObject entries containing name information
// for battle end text and messages. Each battle end text includes localized text for all supported
// languages in the game. The data is loaded directly into the global BTLEND_TEXT variable.
//
// File format: name only data
// Pattern path: "battle/kernel/btlend_txt.bin"
func ReadBattleEndTextWithAllLocalizations() {
	patternPath := "battle/kernel/btlend_txt.bin"
	components.BTLEND_TEXT = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d battle end text entries with all localizations", components.BTLEND_TEXT.GetLength())
}

// ReadMonsterMagic1WithAllLocalizations reads monster magic 1 data from the monmagic1.bin file
// and loads all available localizations for each monster magic entry directly into components.MONMAGIC1.
//
// This function reads LocalizedTextObject entries containing name information
// for monster magic spells and abilities. Each monster magic includes localized text for all supported
// languages in the game. The data is loaded directly into the global MONMAGIC1 variable.
//
// File format: name only data
// Pattern path: "battle/kernel/monmagic1.bin"
func ReadMonsterMagic1WithAllLocalizations() {
	patternPath := "battle/kernel/monmagic1.bin"
	components.MONMAGIC1 = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d monster magic 1 entries with all localizations", components.MONMAGIC1.GetLength())
}

// ReadMonsterMagic2WithAllLocalizations reads monster magic 2 data from the monmagic2.bin file
// and loads all available localizations for each monster magic entry directly into components.MONMAGIC2.
//
// This function reads LocalizedTextObject entries containing name information
// for monster magic spells and abilities. Each monster magic includes localized text for all supported
// languages in the game. The data is loaded directly into the global MONMAGIC2 variable.
//
// File format: name only data
// Pattern path: "battle/kernel/monmagic2.bin"
func ReadMonsterMagic2WithAllLocalizations() {
	patternPath := "battle/kernel/monmagic2.bin"
	components.MONMAGIC2 = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d monster magic 2 entries with all localizations", components.MONMAGIC2.GetLength())
}

// ReadBuildTextWithAllLocalizations reads build text data from the build_txt.bin file
// and loads all available localizations for each build text entry directly into components.BUILD_TEXT.
//
// This function reads LocalizedTextObject entries containing name and description information
// for build-related text and construction messages. Each build text includes localized text for all supported
// languages in the game. The data is loaded directly into the global BUILD_TEXT variable.
//
// File format: name and description data
// Pattern path: "battle/kernel/build_txt.bin"
func ReadBuildTextWithAllLocalizations() {
	patternPath := "battle/kernel/build_txt.bin"
	components.BUILD_TEXT = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d build text entries with all localizations", components.BUILD_TEXT.GetLength())
}

// ReadNameTextWithAllLocalizations reads name text data from the name_txt.bin file
// and loads all available localizations for each name text entry directly into components.NAME_TEXT.
//
// This function reads LocalizedTextObject entries containing name information
// for name text and character names. Each name text includes localized text for all supported
// languages in the game. The data is loaded directly into the global NAME_TEXT variable.
//
// File format: name only data
// Pattern path: "battle/kernel/name_txt.bin"
func ReadNameTextWithAllLocalizations() {
	patternPath := "battle/kernel/name_txt.bin"
	components.NAME_TEXT = ReadNameOnlyDataObjectsWithIlist(patternPath)

	common.LogVerbose("Loaded %d name text entries with all localizations", components.NAME_TEXT.GetLength())
}
