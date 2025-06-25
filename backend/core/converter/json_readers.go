package converter

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"fmt"
)

// ProcessKeyItemsJsonFile reads a JSON file and updates key items data in components.KEY_ITEMS.
//
// This function imports localized text data from "key_items_all_localizations.json" and
// applies the translations directly to the global KEY_ITEMS variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: key_items_all_localizations.json
// Target: components.KEY_ITEMS
//
// Returns: error if import fails or file cannot be read
func ProcessKeyItemsJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"key_items_all_localizations.json",
		components.KEY_ITEMS,
	)
}

// ProcessCommandsJsonFile reads a JSON file and updates commands data in components.COMMANDS.
//
// This function imports localized text data from "commands_all_localizations.json" and
// applies the translations directly to the global COMMANDS variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: commands_all_localizations.json
// Target: components.COMMANDS
//
// Returns: error if import fails or file cannot be read
func ProcessCommandsJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"commands_all_localizations.json",
		components.COMMANDS,
	)
}

// ProcessItemsJsonFile reads a JSON file and updates items data in components.ITEMS.
//
// This function imports localized text data from "items_all_localizations.json" and
// applies the translations directly to the global ITEMS variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: items_all_localizations.json
// Target: components.ITEMS
//
// Returns: error if import fails or file cannot be read
func ProcessItemsJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"items_all_localizations.json",
		components.ITEMS,
	)
}

// ProcessArmsJsonFile reads a JSON file and updates arms text data in components.ARMS_TEXT.
//
// This function imports localized text data from "arms_all_localizations.json" and
// applies the translations directly to the global ARMS_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: arms_all_localizations.json
// Target: components.ARMS_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessArmsJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"arms_all_localizations.json",
		components.ARMS_TEXT,
	)
}

// ProcessBattleTextJsonFile reads a JSON file and updates battle text data in components.BTL_TEXT.
//
// This function imports localized text data from "battle_text_all_localizations.json" and
// applies the translations directly to the global BTL_TEXT variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: battle_text_all_localizations.json
// Target: components.BTL_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessBattleTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"battle_text_all_localizations.json",
		components.BTL_TEXT,
	)
}

// ProcessBattleEndTextJsonFile reads a JSON file and updates battle end text data in components.BTLEND_TEXT.
//
// This function imports localized text data from "battle_end_text_all_localizations.json" and
// applies the translations directly to the global BTLEND_TEXT variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: battle_end_text_all_localizations.json
// Target: components.BTLEND_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessBattleEndTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"battle_end_text_all_localizations.json",
		components.BTLEND_TEXT,
	)
}

// ProcessMonsterMagic1JsonFile reads a JSON file and updates monster magic 1 data in components.MONMAGIC1.
//
// This function imports localized text data from "monster_magic1_all_localizations.json" and
// applies the translations directly to the global MONMAGIC1 variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: monster_magic1_all_localizations.json
// Target: components.MONMAGIC1
//
// Returns: error if import fails or file cannot be read
func ProcessMonsterMagic1JsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"monster_magic1_all_localizations.json",
		components.MONMAGIC1,
	)
}

// ProcessMonsterMagic2JsonFile reads a JSON file and updates monster magic 2 data in components.MONMAGIC2.
//
// This function imports localized text data from "monster_magic2_all_localizations.json" and
// applies the translations directly to the global MONMAGIC2 variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: monster_magic2_all_localizations.json
// Target: components.MONMAGIC2
//
// Returns: error if import fails or file cannot be read
func ProcessMonsterMagic2JsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"monster_magic2_all_localizations.json",
		components.MONMAGIC2,
	)
}

// ProcessBuildTextJsonFile reads a JSON file and updates build text data in components.BUILD_TEXT.
//
// This function imports localized text data from "build_all_localizations.json" and
// applies the translations directly to the global BUILD_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: build_all_localizations.json
// Target: components.BUILD_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessBuildTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"build_all_localizations.json",
		components.BUILD_TEXT,
	)
}

// ProcessConfigTextJsonFile reads a JSON file and updates config text data in components.CONFIG_TEXT.
//
// This function imports localized text data from "config_text_all_localizations.json" and
// applies the translations directly to the global CONFIG_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: config_text_all_localizations.json
// Target: components.CONFIG_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessConfigTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"config_text_all_localizations.json",
		components.CONFIG_TEXT,
	)
}

// ProcessItemCommandsJsonFile reads a JSON file and updates item text data in components.ITEM_TEXT.
//
// This function imports localized text data from "item_commands_all_localizations.json" and
// applies the translations directly to the global ITEM_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: item_commands_all_localizations.json
// Target: components.ITEM_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessItemCommandsJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"item_commands_all_localizations.json",
		components.ITEM_TEXT,
	)
}

// ProcessMainMenuTextJsonFile reads a JSON file and updates main menu text data in components.MMAIN_TEXT.
//
// This function imports localized text data from "main_menu_all_localizations.json" and
// applies the translations directly to the global MMAIN_TEXT variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: main_menu_all_localizations.json
// Target: components.MMAIN_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessMainMenuTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"main_menu_all_localizations.json",
		components.MMAIN_TEXT,
	)
}

// ProcessPlayerRoomTextJsonFile reads a JSON file and updates player room text data in components.PLAYER_ROOM.
//
// This function imports localized text data from "player_room_all_localizations.json" and
// applies the translations directly to the global PLAYER_ROOM variable. The JSON file should
// contain name and description information for all supported languages.
//
// JSON file: player_room_all_localizations.json
// Target: components.PLAYER_ROOM
//
// Returns: error if import fails or file cannot be read
func ProcessPlayerRoomTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"player_room_all_localizations.json",
		components.PLAYER_ROOM,
	)
}

// ProcessNameTextJsonFile reads a JSON file and updates name text data in components.NAME_TEXT.
//
// This function imports localized text data from "names_all_localizations.json" and
// applies the translations directly to the global NAME_TEXT variable. The JSON file should
// contain name information for all supported languages.
//
// JSON file: names_all_localizations.json
// Target: components.NAME_TEXT
//
// Returns: error if import fails or file cannot be read
func ProcessNameTextJsonFile() error {
	return ImportLocalizedDataFromJsonFile(
		"names_all_localizations.json",
		components.NAME_TEXT,
	)
}

// ProcessEventsJsonFile reads a JSON file and updates all events data in components.EVENTS.
//
// This function imports event data from "events_all_localizations.json" and applies
// the translations to all events found in the JSON file. The JSON file should contain
// event string information for all supported languages.
//
// JSON file: events_all_localizations.json
// Target: components.EVENTS (multiple entries)
//
// Returns: error if import fails or file cannot be read
func ProcessEventsJsonFile() error {
	return processEventJsonFile("", false)
}

// ProcessEventJsonFile reads a JSON file and updates a specific event in components.EVENTS.
//
// This function imports event data from "events_all_localizations.json" and applies
// the translations to a single specified event. The JSON file should contain
// event string information for all supported languages.
//
// JSON file: events_all_localizations.json
// Target: components.EVENTS (single entry specified by eventID)
//
// Parameters:
//   - eventID: The ID of the specific event to process (e.g., "ev001", "btl_001")
//
// Returns: error if import fails, file cannot be read, or event is not found
func ProcessEventJsonFile(eventID string) error {
	return processEventJsonFile(eventID, true)
}

// processEventJsonFile is the core function that handles both single and multiple event processing.
// This function centralizes the common logic between ProcessEventsJsonFile and ProcessEventJsonFile
// to avoid code duplication while providing flexibility for different processing modes.
//
// Parameters:
//   - eventID: The specific event ID to process (empty string for all events)
//   - singleEvent: Whether to process only a single event (true) or all events (false)
//
// Returns: error if processing fails
func processEventJsonFile(eventID string, singleEvent bool) error {
	jsonFilePath, err := getEventsJsonFilePath()
	if err != nil {
		return err
	}

	eventDataList, err := loadEventJsonData(jsonFilePath)
	if err != nil {
		return err
	}

	if singleEvent {
		return processSingleEventData(eventDataList, eventID)
	}

	return processAllEventData(eventDataList)
}

// updateEventString updates a single string within an event based on JSON data.
// This function handles the string-level updates including localization processing.
//
// Parameters:
//   - eventFile: The event file object to update
//   - eventString: The string data from JSON
//   - eventID: The event ID (for logging purposes)
//
// Returns: error if string index is invalid or update fails
func updateEventString(eventFile *components.EventFile, eventString EventStringData, eventID string) error {
	stringIndex := eventString.Index

	common.LogVerbose("Processing string %d for event %s", stringIndex, eventID)

	if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
		return fmt.Errorf("string index out of range for event %s: %d", eventID, stringIndex)
	}

	objToEdit := eventFile.Strings[stringIndex]

	common.LogVerbose("Updating event %s[%d] with %d localizations",
		eventID, stringIndex, len(eventString.Text))

	// Apply each localization
	for localization, newString := range eventString.Text {
		if err := updateStringLocalization(objToEdit, localization, newString); err != nil {
			common.LogVerbose("Warning: failed to update localization %s for event %s[%d]: %v",
				localization, eventID, stringIndex, err)
		}
	}

	return nil
}

// updateStringLocalization updates a single localization for a string object.
// This function handles the low-level localization update logic.
//
// Parameters:
//   - objToEdit: The string object to update
//   - localization: The localization code (e.g., "us", "jp")
//   - newString: The new string content
//
// Returns: error if localization is not supported or update fails
func updateStringLocalization(objToEdit *components.LocalizedFieldStringObject, localization, newString string) error {
	if newString == "" {
		return nil // Skip empty strings
	}

	if _, exists := common.SupportedLanguages[localization]; !exists {
		return fmt.Errorf("unsupported localization: %s", localization)
	}

	fieldString := objToEdit.GetLocalizedContent(localization)
	if fieldString == nil {
		return fmt.Errorf("failed to get localized content for %s", localization)
	}

	fieldString.SetRegularString(newString)
	return nil
}
