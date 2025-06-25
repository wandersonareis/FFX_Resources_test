package converter

import (
	"ffxresources/backend/core/components"
	"fmt"
)

// WriteAllKeyItemsData writes key items data to the important.bin file
// for all available localizations. It validates that key items are loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no key items are loaded or if the write operation fails
//
// The function writes all loaded key items from components.KEY_ITEMS to the
// "battle/kernel/important.bin" file path across all localizations using the
// ExportLocalizedTextData function.
func WriteAllKeyItemsData() error {
	if components.KEY_ITEMS == nil || components.KEY_ITEMS.IsEmpty() {
		return fmt.Errorf("no key items loaded")
	}

	pathPattern := "battle/kernel/important.bin"
	return ExportLocalizedTextData(components.KEY_ITEMS, pathPattern)
}

// WriteAllItemsData writes item data to the item.bin file
// for all available localizations. It validates that items are loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no items are loaded or if the write operation fails
//
// The function writes all loaded items from components.ITEMS to the
// "battle/kernel/item.bin" file path across all localizations.
func WriteAllItemsData() error {
	if components.ITEMS == nil || components.ITEMS.IsEmpty() {
		return fmt.Errorf("no items loaded")
	}
	pathPattern := "battle/kernel/item.bin"
	return ExportLocalizedTextData(components.ITEMS, pathPattern)
}

// WriteAllCommandsData writes command data to the command.bin file
// for all available localizations. It validates that commands are loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no commands are loaded or if the write operation fails
//
// The function writes all loaded commands from components.COMMANDS to the
// "battle/kernel/command.bin" file path across all localizations.
func WriteAllCommandsData() error {
	if components.COMMANDS == nil || components.COMMANDS.IsEmpty() {
		return fmt.Errorf("no commands loaded")
	}
	pathPattern := "battle/kernel/command.bin"
	return ExportLocalizedTextData(components.COMMANDS, pathPattern)
}

// WriteAllArmsTextData writes arms text data to the arms_txt.bin file
// for all available localizations. It validates that arms text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no arms text is loaded or if the write operation fails
//
// The function writes all loaded arms text from components.ARMS_TEXT to the
// "battle/kernel/arms_txt.bin" file path across all localizations.
func WriteAllArmsTextData() error {
	if components.ARMS_TEXT == nil || components.ARMS_TEXT.IsEmpty() {
		return fmt.Errorf("no arms text loaded")
	}
	pathPattern := "battle/kernel/arms_txt.bin"
	return ExportLocalizedTextData(components.ARMS_TEXT, pathPattern)
}

// WriteAllMonsterMagic1Data writes monster magic 1 data to the monmagic1.bin file
// for all available localizations. It validates that monster magic 1 data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no monster magic 1 data is loaded or if the write operation fails
//
// The function writes all loaded monster magic 1 data from components.MONMAGIC1 to the
// "battle/kernel/monmagic1.bin" file path across all localizations.
func WriteAllMonsterMagic1Data() error {
	if components.MONMAGIC1 == nil || components.MONMAGIC1.IsEmpty() {
		return fmt.Errorf("no monster magic 1 loaded")
	}
	pathPattern := "battle/kernel/monmagic1.bin"
	return ExportLocalizedTextData(components.MONMAGIC1, pathPattern)
}

// WriteAllMonsterMagic2Data writes monster magic 2 data to the monmagic2.bin file
// for all available localizations. It validates that monster magic 2 data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no monster magic 2 data is loaded or if the write operation fails
//
// The function writes all loaded monster magic 2 data from components.MONMAGIC2 to the
// "battle/kernel/monmagic2.bin" file path across all localizations.
func WriteAllMonsterMagic2Data() error {
	if components.MONMAGIC2 == nil || components.MONMAGIC2.IsEmpty() {
		return fmt.Errorf("no monster magic 2 loaded")
	}
	pathPattern := "battle/kernel/monmagic2.bin"
	return ExportLocalizedTextData(components.MONMAGIC2, pathPattern)
}

// WriteAllBattleTextData writes battle text data to the btl_txt.bin file
// for all available localizations. It validates that battle text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no battle text is loaded or if the write operation fails
//
// The function writes all loaded battle text from components.BTL_TEXT to the
// "battle/kernel/btl_txt.bin" file path across all localizations.
func WriteAllBattleTextData() error {
	if components.BTL_TEXT == nil || components.BTL_TEXT.IsEmpty() {
		return fmt.Errorf("no battle text loaded")
	}
	pathPattern := "battle/kernel/btl_txt.bin"
	return ExportLocalizedTextData(components.BTL_TEXT, pathPattern)
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
// The function writes all loaded battle end text from components.BTLEND_TEXT to the
// "battle/kernel/btlend_txt.bin" file path across all localizations.
func WriteAllBattleEndTextData() error {
	if components.BTLEND_TEXT == nil || components.BTLEND_TEXT.IsEmpty() {
		return fmt.Errorf("no battle end text loaded")
	}
	pathPattern := "battle/kernel/btlend_txt.bin"
	return ExportLocalizedTextData(components.BTLEND_TEXT, pathPattern)
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
// The function writes all loaded build text from components.BUILD_TEXT to the
// "battle/kernel/build_txt.bin" file path across all localizations.
func WriteAllBuildTextData() error {
	if components.BUILD_TEXT == nil || components.BUILD_TEXT.IsEmpty() {
		return fmt.Errorf("no build text loaded")
	}
	pathPattern := "battle/kernel/build_txt.bin"
	return ExportLocalizedTextData(components.BUILD_TEXT, pathPattern)
}

// WriteAllConfigTextData writes config text data to the config_txt.bin file
// for all available localizations. It validates that config text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no config text is loaded or if the write operation fails
//
// The function writes all loaded config text from components.CONFIG_TEXT to the
// "battle/kernel/config_txt.bin" file path across all localizations.
func WriteAllConfigTextData() error {
	if components.CONFIG_TEXT == nil || components.CONFIG_TEXT.IsEmpty() {
		return fmt.Errorf("no config text loaded")
	}
	pathPattern := "battle/kernel/config_txt.bin"
	return ExportLocalizedTextData(components.CONFIG_TEXT, pathPattern)
}

// WriteAllMainMenuTextData writes main menu text data to the mmain_txt.bin file
// for all available localizations. It validates that main menu text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no main menu text is loaded or if the write operation fails
//
// The function writes all loaded main menu text from components.MMAIN_TEXT to the
// "battle/kernel/mmain_txt.bin" file path across all localizations.
func WriteAllMainMenuTextData() error {
	if components.MMAIN_TEXT == nil || components.MMAIN_TEXT.IsEmpty() {
		return fmt.Errorf("no main menu text loaded")
	}
	pathPattern := "battle/kernel/mmain_txt.bin"
	return ExportLocalizedTextData(components.MMAIN_TEXT, pathPattern)
}

// WriteAllItemCommandsData writes item text data to the item_txt.bin file
// for all available localizations. It validates that item text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no item text is loaded or if the write operation fails
//
// The function writes all loaded item text from components.ITEM_TEXT to the
// "battle/kernel/item_txt.bin" file path across all localizations.
func WriteAllItemCommandsData() error {
	if components.ITEM_TEXT == nil || components.ITEM_TEXT.IsEmpty() {
		return fmt.Errorf("no item text loaded")
	}
	pathPattern := "battle/kernel/item_txt.bin"
	return ExportLocalizedTextData(components.ITEM_TEXT, pathPattern)
}

// WriteAllPlayerRoomTextData writes player room text data to the ply_rom.bin file
// for all available localizations. It validates that player room text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no player room text is loaded or if the write operation fails
//
// The function writes all loaded player room text from components.PLAYER_ROOM to the
// "battle/kernel/ply_rom.bin" file path across all localizations.
func WriteAllPlayerRoomTextData() error {
	if components.PLAYER_ROOM == nil || components.PLAYER_ROOM.IsEmpty() {
		return fmt.Errorf("no player room text loaded")
	}
	pathPattern := "battle/kernel/ply_rom.bin"
	return ExportLocalizedTextData(components.PLAYER_ROOM, pathPattern)
}

// WriteAllNameTextData writes name text data to the name_txt.bin file
// for all available localizations. It validates that name text data is loaded before
// attempting to write the data.
//
// Returns:
//   - error: returns an error if no name text is loaded or if the write operation fails
//
// The function writes all loaded name text from components.NAME_TEXT to the
// "battle/kernel/name_txt.bin" file path across all localizations.
func WriteAllNameTextData() error {
	if components.NAME_TEXT == nil || components.NAME_TEXT.IsEmpty() {
		return fmt.Errorf("no name text loaded")
	}
	pathPattern := "battle/kernel/name_txt.bin"
	return ExportLocalizedTextData(components.NAME_TEXT, pathPattern)
}
