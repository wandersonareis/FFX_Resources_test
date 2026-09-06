package objectsfile

import (
	"encoding/json"
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
	"ffxresources/backend/core/converter"
	"ffxresources/backend/datastore"
	"ffxresources/backend/models"
	"fmt"
	"path/filepath"
)

type (
	NameOnlyData struct {
		ID   int               `json:"id"`
		Name map[string]string `json:"name"`
	}

	NameDescriptionData struct {
		NameOnlyData
		Description map[string]string `json:"description"`
	}

	JobTextData struct {
		NameOnlyData
		Description map[string]string `json:"description"`
		Effect      map[string]string `json:"effect"`
	}

	PlateTextData struct {
		NameOnlyData
		Description map[string]string     `json:"description"`
		Abilities   []map[string]string   `json:"abilities"`
		Effect      map[string]string     `json:"effect"`
	}
)

// ImportLocalizedDataFromJsonFile imports localized text data from a JSON file
// and applies the translations to the corresponding LocalizedTextObject entries.
//
// This function reads JSON files containing localized name and description information
// for game objects, items, commands, and other text elements. Each JSON entry includes
// translations for all supported languages in the game, which are then applied to
// update the existing LocalizedTextObject instances.
//
// The function supports both NameDescriptionTextObject and NameOnlyTextObject types,
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
func ImportLocalizedDataFromJsonFile(
	jsonFileName string,
	objectsList components.IList[datastore.IGlobalLocalizedTextObject],
) error {
	jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", common.WithVersionSuffix(jsonFileName))
	if !common.IsPathExists(jsonFilePath) {
		// Fallback to a non-versioned filename for backward compatibility and tests
		// that were written before the per-game version suffix was introduced.
		plainPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", jsonFileName)
		if !common.IsPathExists(plainPath) {
			return fmt.Errorf("JSON file not found: %s", jsonFilePath)
		}
		jsonFilePath = plainPath
	}

	common.LogVerbose("Processing JSON file with IList: %s", jsonFilePath)

	loaded, loadErr := models.LoadDataFile[models.ObjectsFileExport](jsonFilePath)
	var jsonData []NameDescriptionData
	if loadErr != nil {
		// Fallback to a raw JSON array (legacy format without wrapper).
		raw, rawErr := common.ReadFile(jsonFilePath)
		if rawErr != nil {
			common.LogVerbose("Error reading JSON file %s: %v", jsonFilePath, rawErr)
			return rawErr
		}
		if err := json.Unmarshal(raw, &jsonData); err != nil {
			common.LogVerbose("Error parsing JSON file %s: %v", jsonFilePath, err)
			return err
		}
	} else {
		if err := json.Unmarshal(loaded.Strings, &jsonData); err != nil {
			common.LogVerbose("Error parsing strings in JSON file %s: %v", jsonFilePath, err)
			return err
		}
	}

	common.LogVerbose("Found %d objects in JSON file", len(jsonData))

	localizedObjects := extractLocalizedObjects(objectsList)

	if err := updateLocalizedObjectEntries(jsonData, localizedObjects); err != nil {
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

// updateObjectByType updates a localized object based on its type.
// Automatically detects the object type and applies appropriate updates.
//
// Parameters:
//   - jsonEntry: JSON data containing the new values
//   - obj: The localized object to update
//
// Returns: error if object type is not supported
func updateObjectByType(jsonEntry NameDescriptionData, obj datastore.IGlobalLocalizedTextObject) error {
	switch typed := obj.(type) {
	case *NameDescriptionTextObject:
		updateNameEntry(jsonEntry, typed)
		updateDescriptionEntry(jsonEntry, typed)
	case *NameDescriptionTextObjectV2:
		updateNameEntryV2(jsonEntry, typed)
		updateDescriptionEntryV2(jsonEntry, typed)
	case *NameOnlyDataObject:
		updateNameOnlyFromObjectsData(jsonEntry, typed)
	default:
		common.LogVerbose("Object type not recognized for ID %d", jsonEntry.ID)
		return fmt.Errorf("unknown type for ID object %d", jsonEntry.ID)
	}
	return nil
}

// updateLocalizedObjectEntries processes JSON data entries and applies the localized
// text content to the corresponding ILocalizedTextObject instances.
//
// This function handles type detection automatically, supporting both NameDescriptionTextObject
// and NameOnlyTextObject types. It validates object IDs, checks language support, and
// updates the appropriate text fields based on the object type.
//
// Parameters:
//   - itemsData: Slice of NameDescriptionData from JSON file
//   - objects: Slice of ILocalizedTextObject instances to be updated
//
// Returns: error if any update operations fail
func updateLocalizedObjectEntries(itemsData []NameDescriptionData, objects []datastore.IGlobalLocalizedTextObject) error {
	for _, jsonEntry := range itemsData {
		id := jsonEntry.ID

		if !isValidID(id, len(objects)) {
			common.LogVerbose("Object ID without range: %d", id)
			continue
		}

		localizedObj := objects[id]
		if localizedObj == nil {
			common.LogVerbose("Localized object not found: %d", id)
			continue
		}

		common.LogVerbose("Processing localized object %d", id)

		if err := updateObjectByType(jsonEntry, localizedObj); err != nil {
			common.LogVerbose("Error updating object %d: %v", id, err)
			return err
		}
	}
	return nil
}

// updateNameEntry applies name updates to a NameDescriptionTextObject.
//
// Parameters:
//   - sourceData: JSON data containing name translations
//   - targetObject: The object to update
func updateNameEntry(sourceData NameDescriptionData, targetObject *NameDescriptionTextObject) {
	if len(sourceData.Name) == 0 || targetObject.Name == nil {
		common.LogVerbose("No name found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, nameText := range sourceData.Name {
		if nameText == "" {
			common.LogVerbose("Empty name for language %s, ignoring...", languageCode)
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for name: %s", languageCode)
			continue
		}

		updateOrCreateName(targetObject, languageCode, nameText)
	}
}

// updateDescriptionEntry applies description updates to a NameDescriptionTextObject.
//
// Parameters:
//   - sourceData: JSON data containing description translations
//   - targetObject: The object to update
func updateDescriptionEntry(sourceData NameDescriptionData, targetObject *NameDescriptionTextObject) {
	if len(sourceData.Description) == 0 || targetObject.Description == nil {
		common.LogVerbose("No description found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, descriptionText := range sourceData.Description {
		if descriptionText == "" {
			common.LogVerbose("Empty description for language %s, ignoring...", languageCode)
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for description: %s", languageCode)
			continue
		}

		updateOrCreateDescription(targetObject, languageCode, descriptionText)
	}
}

// updateNameEntryV2 applies name updates to a NameDescriptionTextObjectV2.
//
// Parameters:
//   - sourceData: JSON data containing name translations
//   - targetObject: The object to update
func updateNameEntryV2(sourceData NameDescriptionData, targetObject *NameDescriptionTextObjectV2) {
	if len(sourceData.Name) == 0 || targetObject.NameSegment == nil {
		common.LogVerbose("No name found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, nameText := range sourceData.Name {
		if nameText == "" {
			common.LogVerbose("Empty name for language %s, ignoring...", languageCode)
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for name: %s", languageCode)
			continue
		}

		updateOrCreateSegment(targetObject.NameSegment, languageCode, nameText)
	}
}

// updateDescriptionEntryV2 applies description updates to a NameDescriptionTextObjectV2.
//
// Parameters:
//   - sourceData: JSON data containing description translations
//   - targetObject: The object to update
func updateDescriptionEntryV2(sourceData NameDescriptionData, targetObject *NameDescriptionTextObjectV2) {
	if len(sourceData.Description) == 0 || targetObject.DescriptionSegment == nil {
		common.LogVerbose("No description found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, descriptionText := range sourceData.Description {
		if descriptionText == "" {
			common.LogVerbose("Empty description for language %s, ignoring...", languageCode)
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for description: %s", languageCode)
			continue
		}

		updateOrCreateSegment(targetObject.DescriptionSegment, languageCode, descriptionText)
	}
}

// updateNameOnlyFromObjectsData applies name updates to a NameOnlyTextObject.
//
// Parameters:
//   - sourceData: JSON data containing name translations
//   - targetObject: The name-only object to update
func updateNameOnlyFromObjectsData(sourceData NameDescriptionData, targetObject *NameOnlyDataObject) {
	if len(sourceData.Name) == 0 || targetObject.Name == nil {
		common.LogVerbose("No name found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, nameText := range sourceData.Name {
		if nameText == "" {
			common.LogVerbose("Empty name for language %s, ignoring...", languageCode)
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for name: %s", languageCode)
			continue
		}

		updateOrCreateNameOnly(targetObject, languageCode, nameText)
	}
}

// createNewKeyedString creates a new KeyedString with the given text and charset.
//
// Parameters:
//   - text: The string content
//   - charset: The character encoding to use
//
// Returns: A new KeyedString instance with default offset and key values
func createNewKeyedString(text string, charset string) *KeyedString {
	return &KeyedString{
		Charset: charset,
		/* 		Offset:  0,
		   		Key:     0, */
		Segment: models.Segment{Offset: 0, Key: 0},
		Bytes:   converter.StringToBytes(text, charset),
	}
}

// updateOrCreateSegment updates or creates the localized content of a single keyed-string
// segment (name or description), shared between the V1 and V2 name+description objects.
//
// Parameters:
//   - segment: The keyed-string segment to update (Name/Description of either version)
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateSegment(segment datastore.IGlobalLocalizedKeyedStringObject, languageCode, newText string) {
	existingContent := segment.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(newText, charset)
	} else {
		segment.SetLocalizedContent(languageCode, createNewKeyedString(newText, charset))
	}

	common.LogVerbose("Segment updated (%s): %s", languageCode, newText)
}

// updateOrCreateName updates or creates a name entry for a NameDescriptionTextObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateName(target *NameDescriptionTextObject, languageCode, newText string) {
	updateOrCreateSegment(target.Name, languageCode, newText)
}

// updateOrCreateDescription updates or creates a description entry for a NameDescriptionTextObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateDescription(target *NameDescriptionTextObject, languageCode, text string) {
	updateOrCreateSegment(target.Description, languageCode, text)
}

// updateOrCreateNameOnly updates or creates a name entry for a NameOnlyDataObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateNameOnly(target *NameOnlyDataObject, languageCode, text string) {
	existingContent := target.Name.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(text, charset)
	} else {
		target.Name.SetLocalizedContent(languageCode, createNewKeyedString(text, charset))
	}

	common.LogVerbose("Name updated (%s): %s", languageCode, text)
}

// updatePlateEntry applies plate updates to a PlateTextObject.
func updatePlateEntry(sourceData PlateTextData, targetObject *PlateTextObject) {
	if len(sourceData.Name) > 0 && targetObject.NameSegment != nil {
		for languageCode, nameText := range sourceData.Name {
			if nameText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.NameSegment, languageCode, nameText)
		}
	}

	if len(sourceData.Description) > 0 && targetObject.DescriptionSegment != nil {
		for languageCode, descText := range sourceData.Description {
			if descText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.DescriptionSegment, languageCode, descText)
		}
	}

	for i, abMap := range sourceData.Abilities {
		if i >= len(targetObject.AbilitySegments) {
			break
		}
		for languageCode, abilityText := range abMap {
			if abilityText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.AbilitySegments[i], languageCode, abilityText)
		}
	}

	if len(sourceData.Effect) > 0 && targetObject.EffectSegment != nil {
		for languageCode, effectText := range sourceData.Effect {
			if effectText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.EffectSegment, languageCode, effectText)
		}
	}
}

// serializePlateTextToJSON serializes a PlateTextObject to JSON-friendly data.
// Returns a PlateTextData struct with name, description, abilities, and effect for all supported languages.
func serializePlateTextToJSON(obj *PlateTextObject, languageCode string) PlateTextData {
	name := obj.GetName(languageCode)
	description := ""
	if descContent := obj.DescriptionSegment.GetLocalizedContent(languageCode); descContent != nil {
		description = descContent.GetString()
	}

	// AbilityData is a single []byte containing 4 × 30 bytes
	ability := ""
	if len(obj.AbilityData) > 0 {
		ability = string(obj.AbilityData)
	}

	effect := ""
	if effContent := obj.EffectSegment.GetLocalizedContent(languageCode); effContent != nil {
		effect = effContent.GetString()
	}

	return PlateTextData{
		NameOnlyData: NameOnlyData{
			ID: 0,
			Name: map[string]string{
				languageCode: name,
			},
		},
		Description: map[string]string{
			languageCode: description,
		},
		Abilities: []map[string]string{
			{languageCode: ability},
		},
		Effect: map[string]string{
			languageCode: effect,
		},
	}
}

// ImportPlateDataFromJsonFile imports plate data from a JSON file and applies
// translations to the corresponding PlateTextObject entries.
func ImportPlateDataFromJsonFile(jsonFileName string, objectsList components.IList[datastore.IGlobalLocalizedTextObject]) error {
	jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", common.WithVersionSuffix(jsonFileName))
	if !common.IsPathExists(jsonFilePath) {
		plainPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", jsonFileName)
		if !common.IsPathExists(plainPath) {
			return fmt.Errorf("JSON file not found: %s", jsonFilePath)
		}
		jsonFilePath = plainPath
	}

	common.LogVerbose("Processing plate JSON file: %s", jsonFilePath)

	var jsonData []PlateTextData

	loaded, loadErr := models.LoadDataFile[models.ObjectsFileExport](jsonFilePath)
	if loadErr != nil {
		raw, rawErr := common.ReadFile(jsonFilePath)
		if rawErr != nil {
			return rawErr
		}
		if err := json.Unmarshal(raw, &jsonData); err != nil {
			return err
		}
	} else {
		if err := json.Unmarshal(loaded.Strings, &jsonData); err != nil {
			return err
		}
	}

	if objectsList == nil || objectsList.IsEmpty() {
		return fmt.Errorf("no plate objects loaded")
	}

	objects := objectsList.Items()

	for _, jsonEntry := range jsonData {
		id := jsonEntry.ID
		if !isValidID(id, len(objects)) {
			continue
		}

		plateObj, ok := objects[id].(*PlateTextObject)
		if !ok {
			continue
		}

		updatePlateEntry(jsonEntry, plateObj)
	}

	common.LogVerbose("Plate objects updated from JSON: %s", jsonFileName)
	return nil
}
