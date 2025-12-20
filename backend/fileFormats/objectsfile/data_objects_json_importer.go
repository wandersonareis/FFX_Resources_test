package objectsfile

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
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
	jsonFilePath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits", jsonFileName)
	if !common.IsPathExists(jsonFilePath) {
		return fmt.Errorf("JSON file not found: %s", jsonFilePath)
	}

	common.LogVerbose("Processing JSON file with IList: %s", jsonFilePath)

	jsonData, err := common.ReadJsonFile[[]NameDescriptionData](jsonFilePath)
	if err != nil {
		common.LogVerbose("Error reading JSON file %s: %v", jsonFilePath, err)
		return err
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
		Bytes:   components.StringToBytes(text, charset),
	}
}

// updateOrCreateName updates or creates a name entry for a NameDescriptionTextObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateName(target *NameDescriptionTextObject, languageCode, newText string) {
	existingContent := target.Name.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(newText, charset)
	} else {
		target.Name.SetLocalizedContent(languageCode, createNewKeyedString(newText, charset))
	}

	common.LogVerbose("Name updated (%s): %s", languageCode, newText)
}

// updateOrCreateDescription updates or creates a description entry for a NameDescriptionTextObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateDescription(target *NameDescriptionTextObject, languageCode, text string) {
	existingContent := target.Description.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(text, charset)
	} else {
		target.Description.SetLocalizedContent(languageCode, createNewKeyedString(text, charset))
	}

	common.LogVerbose("Description updated (%s): %s", languageCode, text)
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
