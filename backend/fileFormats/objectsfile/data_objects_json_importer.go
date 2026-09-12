package objectsfile

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/datastore"
	"ffxresources/backend/interactions"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
	"strconv"
)

type (
	JSONEntry struct {
		ID                    int                    `json:"id"`
		Name                  map[string]string      `json:"name,omitempty"`
		SimplifiedName        map[string]string      `json:"simplifiedName,omitempty"`
		Description           map[string]string      `json:"description,omitempty"`
		SimplifiedDescription map[string]string      `json:"simplifiedDescription,omitempty"`
		Effect                map[string]string      `json:"effect,omitempty"`
		EffectDescription     map[string]string      `json:"effectDescription,omitempty"`
		Abilities             []map[string]string    `json:"abilities,omitempty"`
		SensorText            map[string]string      `json:"sensorText,omitempty"`
		SimplifiedSensorText  map[string]string      `json:"simplifiedSensorText,omitempty"`
		ScanText              map[string]string      `json:"scanText,omitempty"`
		SimplifiedScanText    map[string]string      `json:"simplifiedScanText,omitempty"`
		Weapons               map[string]WeaponTexts `json:"weapons,omitempty"`
	}

	WeaponTexts struct {
		Name           map[string]string `json:"name"`
		SimplifiedName map[string]string `json:"simplifiedName"`
	}
	NameOnlyData struct {
		ID   int               `json:"id"`
		Name map[string]string `json:"name"`
	}

	NameDescriptionData struct {
		NameOnlyData
		Description map[string]string `json:"description"`
	}
)

// ImportFromJson imports localized text data from a JSON file
// and applies the translations to the corresponding LocalizedTextObject entries.
//
// This function reads JSON files containing localized name and description information
// for game objects, items, commands, and other text elements. Each JSON entry includes
// translations for all supported languages in the game, which are then applied to
// update the existing LocalizedTextObject instances.
//
// The function supports both CommandTextObject and NameOnlyTextObject types,
// automatically detecting the object type and applying the appropriate updates.
//
// File format: JSON array with id, name, and description mappings
// JSON structure: [{"id": 0, "name": {"us": "...", "sp": "..."}, "description": {"us": "...", "sp": "..."}}]
//
// Parameters:
//   - jsonFileName: Name of the JSON file in the edits folder
//   - objectsList: IList containing ILocalizedTextObject entries to be updated
//
// Returns: error if file not found, invalid JSON format, or update failures
func ImportFromJson(
	jsonFileName string,
	objectsList components.IList[datastore.IGlobalLocalizedTextObject],
) error {
	jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", common.WithVersionSuffix(jsonFileName))
	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Processing JSON file: %s", jsonFilePath)

	loaded, err := models.LoadDataFile[models.ObjectsFileExport](jsonFilePath)
	if err != nil {
		common.LogError("Error loading JSON file %s: %v", jsonFilePath, err)
		return err
	}

	var jsonData []JSONEntry
	if err := json.Unmarshal(loaded.Strings, &jsonData); err != nil {
		common.LogError("Error parsing strings in JSON file %s: %v", jsonFilePath, err)
		return err
	}

	common.LogVerbose("Found %d objects in JSON file", len(jsonData))

	localizedObjects := extractLocalizedObjects(objectsList)

	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	if err := updateLocalizedObjectEntries(jsonData, localizedObjects, version); err != nil {
		common.LogVerbose("Error processing JSON file: %v", err)
		return err
	}

	common.LogVerbose("Localized objects updated successfully from JSON file: %s", jsonFileName)
	return nil
}

// extractLocalizedObjects extracts ILocalizedTextObject instances from an IList
// and returns them as a regular slice for processing.
//
// Parameters:
//   - objectsList: IList containing ILocalizedTextObject entries
//
// Returns: Slice of ILocalizedTextObject instances, or nil if the list is empty
func extractLocalizedObjects(objectsList components.IList[datastore.IGlobalLocalizedTextObject]) []datastore.IGlobalLocalizedTextObject {
	if objectsList == nil || objectsList.IsEmpty() {
		return nil
	}

	return objectsList.Items()
}

// isValidID checks if an ID is within valid range for a collection.
//
// Parameters:
//   - id: The ID to validate
//   - length: The size of the collection
//
// Returns: true if ID is valid (>= 0 and < length), false otherwise
func isValidID(id int, length int) bool {
	return id >= 0 && id < length
}

// updateObject applies JSON data to a localized object. Static keyed fields and
// abilities are applied generically via GetKeyedString (absent keys no-op); weapons
// remain a specific mapping due to their per-character nested JSON structure.
func updateObject(jsonEntry JSONEntry, obj datastore.IGlobalLocalizedTextObject, version common.GameVersion) {
	for _, f := range staticFields {
		applyLocalizedText(obj.GetKeyedString(f.key), *f.get(&jsonEntry), version, f.label)
	}
	for i, amap := range jsonEntry.Abilities {
		applyLocalizedText(obj.GetKeyedString(fmt.Sprintf("ability%d", i+1)), amap, version, "ability "+strconv.Itoa(i+1))
	}
	if w, ok := obj.(*WeaponsNameTextObject); ok {
		updateWeaponsEntry(jsonEntry, w, version)
	}
}

// updateLocalizedObjectEntries processes JSON data entries and applies the localized
// text content to the corresponding ILocalizedTextObject instances.
//
// This function handles type detection automatically, supporting both CommandTextObject
// and NameOnlyTextObject types. It validates object IDs, checks language support, and
// updates the appropriate text fields based on the object type.
//
// Parameters:
//   - itemsData: Slice of JSONEntry from JSON file
//   - objects: Slice of ILocalizedTextObject instances to be updated
//
// Returns: error if any update operations fail
func updateLocalizedObjectEntries(itemsData []JSONEntry, objects []datastore.IGlobalLocalizedTextObject, version common.GameVersion) error {
	for _, jsonEntry := range itemsData {
		id := jsonEntry.ID

		if !isValidID(id, len(objects)) {
			common.LogError("Object ID without range: %d", id)
			continue
		}

		localizedObj := objects[id]
		if localizedObj == nil {
			common.LogError("Localized object not found: %d", id)
			continue
		}

		common.LogVerbose("Processing localized object %d", id)

		updateObject(jsonEntry, localizedObj, version)
	}
	return nil
}



// createNewKeyedString creates a new KeyedString with the given text and charset.
//
// Parameters:
//   - text: The string content
//   - charset: The character encoding to use
//   - version: Game version as int (1 = FFX, 2 = FFX-2)
//
// Returns: A new KeyedString instance with default offset and key values
func createNewKeyedString(text string, charset string, version common.GameVersion) *KeyedString {
	return &KeyedString{
		Charset: charset,
		Version: version,
/* 		Offset:  0,
	   		Key:     0, */
		Segment: models.Segment{Offset: 0, Key: 0},
		Bytes:   converter.StringToBytes(text, charset, version),
	}
}

// updateOrCreateSegment updates or creates the localized content of a single keyed-string
// segment (name or description), shared between the V1 and V2 name+description objects.
//
// Parameters:
//   - segment: The keyed-string segment to update (Name/Description of either version)
//   - text: The new text content
//   - languageCode: Language code (e.g., "us", "sp")
func updateOrCreateSegment(segment datastore.IGlobalLocalizedKeyedStringObject, newText, languageCode string, version common.GameVersion) {
	existingContent := segment.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(newText, charset)
	} else {
		segment.SetLocalizedContent(languageCode, createNewKeyedString(newText, charset, version))
	}

	common.LogVerbose("Segment updated (%s): %s", languageCode, newText)
}

func updateWeaponsEntry(sourceData JSONEntry, obj *WeaponsNameTextObject, version common.GameVersion) {
	if len(sourceData.Weapons) == 0 {
		common.LogVerbose("No weapons found, skipping...")
		return
	}

	for _, ref := range weaponRefs {
		texts, ok := sourceData.Weapons[ref.name]
		if !ok {
			continue
		}
		updateWeaponField(obj, ref.key, texts.Name, ref.name, version)
		updateWeaponField(obj, "s"+ref.key, texts.SimplifiedName, ref.name+" simplified", version)
	}
}

func updateWeaponField(obj *WeaponsNameTextObject, key string, fieldTexts map[string]string, label string, version common.GameVersion) {
	if len(fieldTexts) == 0 {
		return
	}
	segment := obj.GetKeyedString(key)
	if segment == nil {
		common.LogVerbose("Weapon segment %q is nil, skipping...", key)
		return
	}

	applyLocalizedText(segment, fieldTexts, version, label)
}
