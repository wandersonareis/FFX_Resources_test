# Converter Package

This package provides functionality for converting between different data formats used in the FFX Resources system.

## Structure

```
backend/core/converter/
├── binary_reader.go   # Core binary file reading functions
├── binary_writer.go   # Core binary file writing functions
├── data_readers.go    # High-level data reading functions (all game data types)
├── data_writers.go    # High-level data writing functions (all game data types)
├── json_importer.go   # JSON import and processing functions
├── json_exporter.go   # JSON export functions for all game data types
├── json_readers.go    # JSON file reading functions that update global variables
├── convenience.go     # High-level convenience wrappers
├── types.go          # Type definitions and data structures
└── README.md         # This documentation
```

## Main Functions

### Core Binary Reading (`binary_reader.go`)

- **`ReadNameOnlyDataObjectsWithIlist(patternPath string)`** - Reads binary files containing name-only data
- **`ReadCommandObjectsWithIlist(patternPath string)`** - Reads binary files containing name and description data
- **`ReadNameOnlyObjects(patternPath string)`** - Legacy function for name-only data (returns slice)
- **`ReadCommandObjects(patternPath string)`** - Legacy function for name and description data (returns slice)

### High-level Data Reading (`data_readers.go`)

- **`ReadCommandsWithAllLocalizations()`** - Reads commands from command.bin → components.COMMANDS
- **`ReadKeyItemsWithAllLocalizations()`** - Reads key items from important.bin → components.KEY_ITEMS
- **`ReadItemsWithAllLocalizations()`** - Reads items from item.bin → components.ITEMS
- **`ReadArmsTextWithAllLocalizations()`** - Reads weapon/armor text from arms_txt.bin → components.ARMS_TEXT
- **`ReadConfigTextWithAllLocalizations()`** - Reads config text from config_txt.bin → components.CONFIG_TEXT
- **`ReadItemTextWithAllLocalizations()`** - Reads item descriptions from item_txt.bin → components.ITEM_TEXT
- **`ReadMainMenuTextWithAllLocalizations()`** - Reads menu text from mmain_txt.bin → components.MMAIN_TEXT
- **`ReadPlayerRomTextWithAllLocalizations()`** - Reads player ROM text from ply_rom.bin → components.PLAYER_ROOM
- **`ReadBattleTextWithAllLocalizations()`** - Reads battle text from btl_txt.bin → components.BTL_TEXT
- **`ReadBattleEndTextWithAllLocalizations()`** - Reads battle end text from btlend_txt.bin → components.BTLEND_TEXT
- **`ReadMonsterMagic1WithAllLocalizations()`** - Reads monster magic 1 from monmagic1.bin → components.MONMAGIC1
- **`ReadMonsterMagic2WithAllLocalizations()`** - Reads monster magic 2 from monmagic2.bin → components.MONMAGIC2
- **`ReadBuildTextWithAllLocalizations()`** - Reads build text from build_txt.bin → components.BUILD_TEXT
- **`ReadNameTextWithAllLocalizations()`** - Reads name text from name_txt.bin → components.NAME_TEXT

*Note: These functions load data directly into global variables in the components package and do not return values.*

### Core Binary Writing (`binary_writer.go`)

- **`ExportLocalizedTextData(objects, pathPattern)`** - Main function for exporting IList data to binary files
- **`writeLocalizedDataObjectsInAllLocalizations(pathPattern, objects, startIndex, endIndex)`** - Core writing logic
- **`convertLocalizedDataToBytes(objects, from, to, localization)`** - Converts objects to binary format

### High-level Data Writing (`data_writers.go`)

- **`WriteAllKeyItemsData()`** - Writes key items to important.bin
- **`WriteAllItemsData()`** - Writes items to item.bin
- **`WriteAllCommandsData()`** - Writes commands to command.bin
- **`WriteAllArmsTextData()`** - Writes arms text to arms_txt.bin
- **`WriteAllMonsterMagic1Data()`** - Writes monster magic 1 to monmagic1.bin
- **`WriteAllMonsterMagic2Data()`** - Writes monster magic 2 to monmagic2.bin
- **`WriteAllBattleTextData()`** - Writes battle text to btl_txt.bin
- **`WriteAllBattleEndTextData()`** - Writes battle end text to btlend_txt.bin ⚠️ *Contains data structure bug: description field points to name field parts*
- **`WriteAllBuildTextData()`** - Writes build text to build_txt.bin ⚠️ *Contains data structure bug: description field points to name field parts*
- **`WriteAllConfigTextData()`** - Writes config text to config_txt.bin
- **`WriteAllMainMenuTextData()`** - Writes main menu text to mmain_txt.bin
- **`WriteAllItemDescriptionsData()`** - Writes item descriptions to item_txt.bin
- **`WriteAllPlayerRoomTextData()`** - Writes player room text to ply_rom.bin
- **`WriteAllNameTextData()`** - Writes name text to name_txt.bin

### JSON Export (`json_exporter.go`)

- **`ExportKeyItemsToJSON()`** - Exports key items from components.KEY_ITEMS to JSON
- **`ExportCommandsToJSON()`** - Exports commands from components.COMMANDS to JSON
- **`ExportItemsToJSON()`** - Exports items from components.ITEMS to JSON
- **`ExportArmsToJSON()`** - Exports arms text from components.ARMS_TEXT to JSON
- **`ExportConfigToJSON()`** - Exports config text from components.CONFIG_TEXT to JSON
- **`ExportItemDescriptionsToJSON()`** - Exports item descriptions from components.ITEM_TEXT to JSON
- **`ExportMainMenuToJSON()`** - Exports main menu text from components.MMAIN_TEXT to JSON
- **`ExportPlayerRoomToJSON()`** - Exports player room text from components.PLAYER_ROOM to JSON
- **`ExportBuildToJSON()`** - Exports build text from components.BUILD_TEXT to JSON
- **`ExportBattleToJSON()`** - Exports battle text from components.BTL_TEXT to JSON
- **`ExportBattleEndToJSON()`** - Exports battle end text from components.BTLEND_TEXT to JSON
- **`ExportMonsterMagic1ToJSON()`** - Exports monster magic 1 from components.MONMAGIC1 to JSON
- **`ExportMonsterMagic2ToJSON()`** - Exports monster magic 2 from components.MONMAGIC2 to JSON
- **`ExportNameToJSON()`** - Exports name text from components.NAME_TEXT to JSON

*Note: These functions export data directly from global variables and generate JSON files in the edits folder.*

### JSON Readers (`json_readers.go`)

- **`ProcessKeyItemsJsonFile()`** - Reads key_items_all_localizations.json → components.KEY_ITEMS
- **`ProcessCommandsJsonFile()`** - Reads commands_all_localizations.json → components.COMMANDS
- **`ProcessItemsJsonFile()`** - Reads items_all_localizations.json → components.ITEMS
- **`ProcessArmsJsonFile()`** - Reads arms_txt_all_localizations.json → components.ARMS_TEXT
- **`ProcessBattleTextJsonFile()`** - Reads btl_txt_all_localizations.json → components.BTL_TEXT
- **`ProcessBattleEndTextJsonFile()`** - Reads btlend_txt_all_localizations.json → components.BTLEND_TEXT
- **`ProcessMonsterMagic1JsonFile()`** - Reads monmagic1_all_localizations.json → components.MONMAGIC1
- **`ProcessMonsterMagic2JsonFile()`** - Reads monmagic2_all_localizations.json → components.MONMAGIC2
- **`ProcessBuildTextJsonFile()`** - Reads build_all_localizations.json → components.BUILD_TEXT
- **`ProcessConfigTextJsonFile()`** - Reads config_all_localizations.json → components.CONFIG_TEXT
- **`ProcessItemTextJsonFile()`** - Reads item_descriptions_all_localizations.json → components.ITEM_TEXT
- **`ProcessMainMenuTextJsonFile()`** - Reads main_menu_all_localizations.json → components.MMAIN_TEXT
- **`ProcessPlayerRoomTextJsonFile()`** - Reads player_room_all_localizations.json → components.PLAYER_ROOM
- **`ProcessNameTextJsonFile()`** - Reads names_all_localizations.json → components.NAME_TEXT

*Note: These functions read JSON files and update global variables directly.*

### JSON Import (`json_importer.go`)

- **`ImportLocalizedDataFromJsonFile(jsonFileName, objectsList)`** - Main function for importing JSON localization data
- **`extractLocalizedObjects(objectsList)`** - Helper to extract objects from IList
- **`updateLocalizedObjectEntries(itemsData, objects)`** - Updates localized text entries

### Convenience Functions (`convenience.go`)

**JSON Import Functions:**

- **`ProcessCommandJsonFile(commandsList, jsonFileName)`** - Process command JSON files
- **`ProcessKeyItemsJsonFile()`** - Process key items JSON
- **`ProcessCommandsJsonFile()`** - Process commands JSON
- **`ProcessItemsJsonFile()`** - Process items JSON
- And more specific processors for different game data types

**Binary Export Functions:**

- **`WriteAllKeyItemsData()`** - Write key items to binary
- **`WriteAllItemsData()`** - Write items to binary
- **`WriteAllCommandsData()`** - Write commands to binary
- **`WriteAllArmsTextData()`** - Write arms text to binary
- And more specific writers for different game data types

## Usage

### Direct Global Variable Population

The high-level data reading functions in `data_readers.go` automatically populate global variables in the `components` package:

```go
import "ffxresources/backend/core/converter"

// Load all game data into global variables
converter.ReadCommandsWithAllLocalizations()      // → components.COMMANDS
converter.ReadKeyItemsWithAllLocalizations()      // → components.KEY_ITEMS  
converter.ReadItemsWithAllLocalizations()         // → components.ITEMS
converter.ReadArmsTextWithAllLocalizations()      // → components.ARMS_TEXT
converter.ReadConfigTextWithAllLocalizations()    // → components.CONFIG_TEXT
converter.ReadItemTextWithAllLocalizations()      // → components.ITEM_TEXT
converter.ReadMainMenuTextWithAllLocalizations()  // → components.MMAIN_TEXT
converter.ReadPlayerRomTextWithAllLocalizations() // → components.PLAYER_ROOM
converter.ReadBattleTextWithAllLocalizations()    // → components.BTL_TEXT
converter.ReadBattleEndTextWithAllLocalizations() // → components.BTLEND_TEXT
converter.ReadMonsterMagic1WithAllLocalizations() // → components.MONMAGIC1
converter.ReadMonsterMagic2WithAllLocalizations() // → components.MONMAGIC2
converter.ReadBuildTextWithAllLocalizations()     // → components.BUILD_TEXT
converter.ReadNameTextWithAllLocalizations()      // → components.NAME_TEXT

// Access data through components package
command := components.GetCommand(0)
keyItem := components.GetKeyItem(0xA000)
```

### JSON Export Operations

Export loaded game data to JSON files for editing:

```go
import "ffxresources/backend/core/converter"

// First load the data
converter.ReadKeyItemsWithAllLocalizations()
converter.ReadCommandsWithAllLocalizations()
converter.ReadItemsWithAllLocalizations()

// Then export to JSON files
err := converter.ExportKeyItemsToJSON()      // → key_items_all_localizations.json
err = converter.ExportCommandsToJSON()       // → commands_all_localizations.json
err = converter.ExportItemsToJSON()          // → items_all_localizations.json
err = converter.ExportArmsToJSON()           // → arms_all_localizations.json
err = converter.ExportConfigToJSON()         // → config_all_localizations.json
err = converter.ExportBattleToJSON()         // → battle_all_localizations.json
err = converter.ExportMonsterMagic1ToJSON()  // → monster_magic1_all_localizations.json

if err != nil {
    log.Printf("Export failed: %v", err)
}
```

### JSON Import Operations

Import edited JSON files back into the game data:

```go
// Import modified JSON data back into memory
err := converter.ImportLocalizedDataFromJsonFile("key_items_all_localizations.json", components.KEY_ITEMS)
if err != nil {
    log.Printf("Import failed: %v", err)
}

### Low-level Binary Operations Examples

### Reading Binary Data

```go
import "ffxresources/backend/core/converter"

// Read name and description objects
commands := converter.ReadCommandObjectsWithIlist("battle/kernel/command.bin")

// Read name-only objects
battleText := converter.ReadNameOnlyDataObjectsWithIlist("battle/kernel/btl_txt.bin")
```

### Importing JSON Data

```go
import "ffxresources/backend/core/converter"

// Import from JSON file
err := converter.ImportLocalizedDataFromJsonFile("commands_all_localizations.json", commandsList)
if err != nil {
    log.Printf("Error importing JSON: %v", err)
}

// Use convenience functions
err = converter.ProcessCommandsJsonFile()
```

### Exporting Binary Data

```go
import "ffxresources/backend/core/converter"

// Export specific data types to binary
err := converter.WriteAllKeyItemsData()
if err != nil {
    log.Printf("Error writing key items: %v", err)
}

// Use convenience functions
err = converter.WriteAllCommandsData()
```

### Direct Binary Operations

```go
import "ffxresources/backend/core/converter"

// Export any IList to binary files
err := converter.ExportLocalizedTextData(myObjectsList, "battle/kernel/mydata.bin")
if err != nil {
    log.Printf("Error exporting data: %v", err)
}
```

### JSON Data Format

## JSON File Formats

### Name and Description Data Format

The expected JSON format for name and description data (items, commands, key items, etc.):

```json
[
  {
    "id": 0,
    "name": {
      "en": "Attack",
      "pt": "Atacar",
      "ja": "攻撃"
    },
    "description": {
      "en": "Physical attack",
      "pt": "Ataque físico", 
      "ja": "物理攻撃"
    }
  }
]
```

### Name Only Data Format

The expected JSON format for name-only data (battle text, monster magic, etc.):

```json
[
  {
    "id": 0,
    "name": {
      "en": "Fire",
      "pt": "Fogo",
      "ja": "ファイア"
    }
  }
]
```

## Migration from Reader Package

The functions previously in `backend/core/reader/` have been moved here with the same functionality:

**Core Reading Functions:**

- `reader.ReadCommandObjectsWithIlist()` → `converter.ReadCommandObjectsWithIlist()`
- `reader.ReadNameOnlyDataObjectsWithIlist()` → `converter.ReadNameOnlyDataObjectsWithIlist()`

**High-level Reading Functions:**

*Note: Function names have been standardized from "WithAllLocations" to "WithAllLocalizations" for consistency.*
*These functions now load data directly into global variables and do not return values.*

- `reader.ReadCommandsWithAllLocalizations()` → `converter.ReadCommandsWithAllLocalizations()` → components.COMMANDS
- `reader.ReadKeyItemsWithAllLocations()` → `converter.ReadKeyItemsWithAllLocalizations()` → components.KEY_ITEMS
- `reader.ReadItemsWithAllLocalizations()` → `converter.ReadItemsWithAllLocalizations()` → components.ITEMS
- `reader.ReadArmsTextWithAllLocations()` → `converter.ReadArmsTextWithAllLocalizations()` → components.ARMS_TEXT
- `reader.ReadConfigTextWithAllLocations()` → `converter.ReadConfigTextWithAllLocalizations()` → components.CONFIG_TEXT
- `reader.ReadItemTextWithAllLocations()` → `converter.ReadItemTextWithAllLocalizations()` → components.ITEM_TEXT
- `reader.ReadMainMenuTextWithAllLocations()` → `converter.ReadMainMenuTextWithAllLocalizations()` → components.MMAIN_TEXT
- `reader.ReadPlayerRomTextWithAllLocations()` → `converter.ReadPlayerRomTextWithAllLocalizations()` → components.PLAYER_ROOM
- `reader.ReadBattleTextWithAllLocations()` → `converter.ReadBattleTextWithAllLocalizations()` → components.BTL_TEXT
- `reader.ReadBattleEndTextWithAllLocations()` → `converter.ReadBattleEndTextWithAllLocalizations()` → components.BTLEND_TEXT
- `reader.ReadMonsterMagic1WithAllLocations()` → `converter.ReadMonsterMagic1WithAllLocalizations()` → components.MONMAGIC1
- `reader.ReadMonsterMagic2WithAllLocations()` → `converter.ReadMonsterMagic2WithAllLocalizations()` → components.MONMAGIC2
- `reader.ReadBuildTextWithAllLocations()` → `converter.ReadBuildTextWithAllLocalizations()` → components.BUILD_TEXT
- `reader.ReadNameTextWithAllLocations()` → `converter.ReadNameTextWithAllLocalizations()` → components.NAME_TEXT

**Import/Export Functions:**

- `reader.importLocalizedDataFromJsonFile()` → `converter.ImportLocalizedDataFromJsonFile()`
- `reader.exportLocalizedTextData()` → `converter.ExportLocalizedTextData()`

**JSON Export Functions (new in converter package):**

- `writer.ExportKeyItemsTextToJSON()` → `converter.ExportKeyItemsToJSON()`
- `writer.ExportCommandsTextToJSON()` → `converter.ExportCommandsToJSON()`
- `writer.ExportArmsTextToJSON()` → `converter.ExportArmsToJSON()`
- `writer.ExportConfigTextToJSON()` → `converter.ExportConfigToJSON()`
- `writer.ExportItemsTextToJSON()` → `converter.ExportItemsToJSON()`
- `writer.ExportItemTextToJSON()` → `converter.ExportItemDescriptionsToJSON()`
- `writer.ExportMainMenuTextToJSON()` → `converter.ExportMainMenuToJSON()`
- `writer.ExportPlayerRoomTextToJSON()` → `converter.ExportPlayerRoomToJSON()`

*Note: "Text" has been removed from function names for consistency and clarity.*

### Write Functions Migration
- `reader.WriteCommandDataForAllLocalizations()` → `converter.WriteAllCommandsData()`
- `reader.WriteKeyItemsDataForAllLocalizations()` → `converter.WriteAllKeyItemsData()`
- `reader.WriteItemsDataForAllLocalizations()` → `converter.WriteAllItemsData()`
- All other `Write??DataForAllLocalizations()` functions have been renamed to `WriteAll??Data()` for consistency

## Design Principles

1. **Single Responsibility** - Each file handles a specific type of conversion
2. **Type Safety** - Uses Go generics for type-safe operations
3. **Backward Compatibility** - Legacy functions maintained for existing code
4. **Extensibility** - Easy to add new format converters (XML, YAML, etc.)
5. **Error Handling** - Comprehensive error reporting and logging

## Future Extensions

This package is designed to be easily extended for additional conversion formats:

- XML import/export
- YAML support
- Binary format writers
- Custom serialization formats
