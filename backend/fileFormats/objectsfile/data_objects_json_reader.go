package objectsfile

import "ffxresources/backend/datastore"

// ProcessKeyItemsJsonFile reads a JSON file and updates key items data in KEY_ITEMS.
//
// This function imports localized text data from "key_items_all_localizations.json" and
// applies the translations directly to the KEY_ITEMS variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: key_items_all_localizations.json
// Target: KEY_ITEMS
//
// Returns: error if import fails or file cannot be read
func ProcessKeyItemsJsonFile() error {
	return ImportFromJson(
		"key_items_all_localizations.json",
		datastore.KeyItems,
	)
}

// ProcessCommandsJsonFile reads a JSON file and updates commands data in COMMANDS.
//
// This function imports localized text data from "commands_all_localizations.json" and
// applies the translations directly to the COMMANDS variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: commands_all_localizations.json
// Target: COMMANDS
//
// Returns: error if import fails or file cannot be read
func ProcessCommandsJsonFile() error {
	return ImportFromJson(
		"commands_all_localizations.json",
		datastore.Commands,
	)
}

// ProcessItemsJsonFile reads a JSON file and updates items data in globaldata.ITEMS.
//
// This function imports localized text data from "items_all_localizations.json" and
// applies the translations directly to the ITEMS variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: items_all_localizations.json
// Target: ITEMS
//
// Returns: error if import fails or file cannot be read
func ProcessItemsJsonFile() error {
	return ImportFromJson(
		"items_all_localizations.json",
		datastore.Items,
	)
}

// ProcessArmsJsonFile reads a JSON file and updates arms text data in ARMS_TEXT.
//
// This function imports localized text data from "arms_all_localizations.json" and
// applies the translations directly to the ARMS_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: arms_all_localizations.json
// Target: ARMS_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessArmsJsonFile() error {
	return ImportFromJson(
		"arms_all_localizations.json",
		datastore.ArmsTxt,
	)
}

// ProcessBattleTextJsonFile reads a JSON file and updates battle text data in BTL_TEXT.
//
// This function imports localized text data from "battle_text_all_localizations.json" and
// applies the translations directly to the BTL_TEXT variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: battle_text_all_localizations.json
// Target: BTL_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessBattleTextJsonFile() error {
	return ImportFromJson(
		"battle_text_all_localizations.json",
		datastore.BattleTxt,
	)
}

// ProcessBattleEndTextJsonFile reads a JSON file and updates battle end text data in BTLEND_TEXT.
//
// This function imports localized text data from "battle_end_text_all_localizations.json" and
// applies the translations directly to the global BTLEND_TEXT variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: battle_end_text_all_localizations.json
// Target: BTLEND_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessBattleEndTextJsonFile() error {
	return ImportFromJson(
		"battle_end_text_all_localizations.json",
		datastore.BattleEndTxt,
	)
}

// ProcessMonsterMagic1JsonFile reads a JSON file and updates monster magic 1 data in MONMAGIC1.
//
// This function imports localized text data from "monster_magic1_all_localizations.json" and
// applies the translations directly to the global MONMAGIC1 variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: monster_magic1_all_localizations.json
// Target: MONMAGIC1
//
// Returns: error if import fails or file cannot be read
func ProcessMonsterMagic1JsonFile() error {
	return ImportFromJson(
		"monster_magic1_all_localizations.json",
		MONMAGIC1,
	)
}

// ProcessMonsterMagic2JsonFile reads a JSON file and updates monster magic 2 data in MONMAGIC2.
//
// This function imports localized text data from "monster_magic2_all_localizations.json" and
// applies the translations directly to the global MONMAGIC2 variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: monster_magic2_all_localizations.json
// Target: MONMAGIC2
//
// Returns: error if import fails or file cannot be read
func ProcessMonsterMagic2JsonFile() error {
	return ImportFromJson(
		"monster_magic2_all_localizations.json",
		MONMAGIC2,
	)
}

// ProcessBuildTextJsonFile reads a JSON file and updates build text data in BUILD_TEXT.
//
// This function imports localized text data from "build_all_localizations.json" and
// applies the translations directly to the global BUILD_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: build_all_localizations.json
// Target: BUILD_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessBuildTextJsonFile() error {
	return ImportFromJson(
		"build_all_localizations.json",
		BUILD_TEXT,
	)
}

// ProcessConfigTextJsonFile reads a JSON file and updates config text data in CONFIG_TEXT.
//
// This function imports localized text data from "config_text_all_localizations.json" and
// applies the translations directly to the global CONFIG_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: config_text_all_localizations.json
// Target: CONFIG_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessConfigTextJsonFile() error {
	return ImportFromJson(
		"config_text_all_localizations.json",
		CONFIG_TEXT,
	)
}

// ProcessItemCommandsJsonFile reads a JSON file and updates item text data in ITEM_TEXT.
//
// This function imports localized text data from "item_commands_all_localizations.json" and
// applies the translations directly to the global ITEM_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: item_commands_all_localizations.json
// Target: ITEM_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessItemCommandsJsonFile() error {
	return ImportFromJson(
		"item_commands_all_localizations.json",
		ITEM_TEXT,
	)
}

// ProcessMainMenuTextJsonFile reads a JSON file and updates main menu text data in MMAIN_TEXT.
//
// This function imports localized text data from "main_menu_all_localizations.json" and
// applies the translations directly to the global MMAIN_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: main_menu_all_localizations.json
// Target: MMAIN_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessMainMenuTextJsonFile() error {
	return ImportFromJson(
		"main_menu_all_localizations.json",
		MMAIN_TEXT,
	)
}

// ProcessPlayerRoomTextJsonFile reads a JSON file and updates player room text data in PLAYER_ROOM.
//
// This function imports localized text data from "player_room_all_localizations.json" and
// applies the translations directly to the global PLAYER_ROOM variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: player_room_all_localizations.json
// Target: PLAYER_ROOM
//
// Returns: error if import fails or file cannot be read
func ProcessPlayerRoomTextJsonFile() error {
	return ImportFromJson(
		"player_room_all_localizations.json",
		PLAYER_ROOM,
	)
}

// ProcessNameTextJsonFile reads a JSON file and updates name text data in NAME_TEXT.
//
// This function imports localized text data from "names_all_localizations.json" and
// applies the translations directly to the global NAME_TEXT variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: names_all_localizations.json
// Target: NAME_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessNameTextJsonFile() error {
	return ImportFromJson(
		"names_all_localizations.json",
		NAME_TEXT,
	)
}

// ProcessPlateJsonFile reads a JSON file and updates plate data in PLATE.
//
// This function imports localized text data from "plate_all_localizations.json" and
// applies the translations directly to the PLATE variable. The JSON file should
// contain name, description, abilities, and effect information for all supported languages.
//
// JSON file: plate_all_localizations.json
// Target: PLATE
//
// Returns: error if import fails or file cannot be read
func ProcessPlateJsonFile() error {
	return ImportPlateDataFromJsonFile(
		"plate_all_localizations.json",
		PLATE,
	)
}
