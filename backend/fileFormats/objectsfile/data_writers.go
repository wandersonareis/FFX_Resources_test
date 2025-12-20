package objectsfile

import (
	"ffxresources/backend/datastore"
	"fmt"
)

// WriteAllKeyItemsData writes key items data to the important.bin file
// for all available localizations. It validates that key items are loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no key items are loaded or if the write operation fails
//
// The function writes all loaded key items from KEY_ITEMS to the
// "battle/kernel/important.bin" file path across all localizations using the
// ExportLocalizedTextData function.
func WriteAllKeyItemsData() error {
	if datastore.KeyItems.IsEmpty() {
		return fmt.Errorf("no key items loaded")
	}

	pathPattern := "battle/kernel/important.bin"
	return ExportLocalizedTextData(datastore.KeyItems, pathPattern)
}

// WriteAllItemsData writes item data to the item.bin file
// for all available localizations. It validates that items are loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no items are loaded or if the write operation fails
//
// The function writes all loaded items from globaldata.ITEMS to the
// "battle/kernel/item.bin" file path across all localizations.
func WriteAllItemsData() error {
	if datastore.Items.IsEmpty() {
		return fmt.Errorf("no items loaded")
	}
	pathPattern := "battle/kernel/item.bin"
	return ExportLocalizedTextData(datastore.Items, pathPattern)
}

// WriteAllCommandsData writes command data to the command.bin file
// for all available localizations. It validates that commands are loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no commands are loaded or if the write operation fails
//
// The function writes all loaded commands from globaldata.COMMANDS to the
// "battle/kernel/command.bin" file path across all localizations.
func WriteAllCommandsData() error {
	if datastore.Commands.IsEmpty() {
		return fmt.Errorf("no commands loaded")
	}
	pathPattern := "battle/kernel/command.bin"
	return ExportLocalizedTextData(datastore.Commands, pathPattern)
}

// WriteAllArmsTextData writes arms text data to the arms_txt.bin file
// for all available localizations. It validates that arms text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no arms text is loaded or if the write operation fails
//
// The function writes all loaded arms text from ARMS_TEXT to the
// "battle/kernel/arms_txt.bin" file path across all localizations.
func WriteAllArmsTextData() error {
	if datastore.ArmsTxt.IsEmpty() {
		return fmt.Errorf("no arms text loaded")
	}
	pathPattern := "battle/kernel/arms_txt.bin"
	return ExportLocalizedTextData(datastore.ArmsTxt, pathPattern)
}

// WriteAllMonsterMagic1Data writes monster magic 1 data to the monmagic1.bin file
// for all available localizations. It validates that monster magic 1 data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no monster magic 1 data is loaded or if the write operation fails
//
// The function writes all loaded monster magic 1 data from MONMAGIC1 to the
// "battle/kernel/monmagic1.bin" file path across all localizations.
func WriteAllMonsterMagic1Data() error {
	if MONMAGIC1 == nil || MONMAGIC1.IsEmpty() {
		return fmt.Errorf("no monster magic 1 loaded")
	}
	pathPattern := "battle/kernel/monmagic1.bin"
	return ExportLocalizedTextData(MONMAGIC1, pathPattern)
}

// WriteAllMonsterMagic2Data writes monster magic 2 data to the monmagic2.bin file
// for all available localizations. It validates that monster magic 2 data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no monster magic 2 data is loaded or if the write operation fails
//
// The function writes all loaded monster magic 2 data from MONMAGIC2 to the
// "battle/kernel/monmagic2.bin" file path across all localizations.
func WriteAllMonsterMagic2Data() error {
	if MONMAGIC2 == nil || MONMAGIC2.IsEmpty() {
		return fmt.Errorf("no monster magic 2 loaded")
	}
	pathPattern := "battle/kernel/monmagic2.bin"
	return ExportLocalizedTextData(MONMAGIC2, pathPattern)
}

// WriteAllBattleTextData writes battle text data to the btl_txt.bin file
// for all available localizations. It validates that battle text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no battle text is loaded or if the write operation fails
//
// The function writes all loaded battle text from BTL_TEXT to the
// "battle/kernel/btl_txt.bin" file path across all localizations.
func WriteAllBattleTextData() error {
	if datastore.BattleTxt.IsEmpty() {
		return fmt.Errorf("no battle text loaded")
	}
	pathPattern := "battle/kernel/btl_txt.bin"
	return ExportLocalizedTextData(datastore.BattleTxt, pathPattern)
}

// WriteAllBattleEndTextData writes battle end text data to the btlend_txt.bin file
// for all available localizations. It validates that battle end text data is loaded before
// attempting to write the data.
//
// WARNING: This file contains name/description data where the description field incorrectly
// points to parts of the name field. This appears to be a bug in the original data structure
// and should ideally contain only name data. The file will be written in name-only format.
//
// Returns:
//   - error: returns an error if no battle end text is loaded or if the write operation fails
//
// The function writes all loaded battle end text from BTLEND_TEXT to the
// "battle/kernel/btlend_txt.bin" file path across all localizations.
func WriteAllBattleEndTextData() error {
	if datastore.BattleEndTxt.IsEmpty() {
		return fmt.Errorf("no battle end text loaded")
	}
	pathPattern := "battle/kernel/btlend_txt.bin"
	return ExportLocalizedTextData(datastore.BattleEndTxt, pathPattern)
}

// WriteAllBuildTextData writes build text data to the build_txt.bin file
// for all available localizations. It validates that build text data is loaded before
// attempting to write the data.
//
// WARNING: This file contains name/description data where the description field incorrectly
// points to parts of the name field. This appears to be a bug in the original data structure
// and should ideally contain only name data. The file will be written in name-only format.
//
// Returns:
//   - error: returns an error if no build text is loaded or if the write operation fails
//
// The function writes all loaded build text from BUILD_TEXT to the
// "battle/kernel/build_txt.bin" file path across all localizations.
func WriteAllBuildTextData() error {
	if BUILD_TEXT == nil || BUILD_TEXT.IsEmpty() {
		return fmt.Errorf("no build text loaded")
	}
	pathPattern := "battle/kernel/build_txt.bin"
	return ExportLocalizedTextData(BUILD_TEXT, pathPattern)
}

// WriteAllConfigTextData writes config text data to the config_txt.bin file
// for all available localizations. It validates that config text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no config text is loaded or if the write operation fails
//
// The function writes all loaded config text from CONFIG_TEXT to the
// "battle/kernel/config_txt.bin" file path across all localizations.
func WriteAllConfigTextData() error {
	if CONFIG_TEXT == nil || CONFIG_TEXT.IsEmpty() {
		return fmt.Errorf("no config text loaded")
	}
	pathPattern := "battle/kernel/config_txt.bin"
	return ExportLocalizedTextData(CONFIG_TEXT, pathPattern)
}

// WriteAllMainMenuTextData writes main menu text data to the mmain_txt.bin file
// for all available localizations. It validates that main menu text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no main menu text is loaded or if the write operation fails
//
// The function writes all loaded main menu text from MMAIN_TEXT to the
// "battle/kernel/mmain_txt.bin" file path across all localizations.
func WriteAllMainMenuTextData() error {
	if MMAIN_TEXT == nil || MMAIN_TEXT.IsEmpty() {
		return fmt.Errorf("no main menu text loaded")
	}
	pathPattern := "battle/kernel/mmain_txt.bin"
	return ExportLocalizedTextData(MMAIN_TEXT, pathPattern)
}

// WriteAllItemCommandsData writes item text data to the item_txt.bin file
// for all available localizations. It validates that item text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no item text is loaded or if the write operation fails
//
// The function writes all loaded item text from ITEM_TEXT to the
// "battle/kernel/item_txt.bin" file path across all localizations.
func WriteAllItemCommandsData() error {
	if ITEM_TEXT == nil || ITEM_TEXT.IsEmpty() {
		return fmt.Errorf("no item text loaded")
	}
	pathPattern := "battle/kernel/item_txt.bin"
	return ExportLocalizedTextData(ITEM_TEXT, pathPattern)
}

// WriteAllPlayerRoomTextData writes player room text data to the ply_rom.bin file
// for all available localizations. It validates that player room text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no player room text is loaded or if the write operation fails
//
// The function writes all loaded player room text from PLAYER_ROOM to the
// "battle/kernel/ply_rom.bin" file path across all localizations.
func WriteAllPlayerRoomTextData() error {
	if PLAYER_ROOM == nil || PLAYER_ROOM.IsEmpty() {
		return fmt.Errorf("no player room text loaded")
	}
	pathPattern := "battle/kernel/ply_rom.bin"
	return ExportLocalizedTextData(PLAYER_ROOM, pathPattern)
}

// WriteAllNameTextData writes name text data to the name_txt.bin file
// for all available localizations. It validates that name text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no name text is loaded or if the write operation fails
//
// The function writes all loaded name text from NAME_TEXT to the
// "battle/kernel/name_txt.bin" file path across all localizations.
func WriteAllNameTextData() error {
	if NAME_TEXT == nil || NAME_TEXT.IsEmpty() {
		return fmt.Errorf("no name text loaded")
	}
	pathPattern := "battle/kernel/name_txt.bin"
	return ExportLocalizedTextData(NAME_TEXT, pathPattern)
}
