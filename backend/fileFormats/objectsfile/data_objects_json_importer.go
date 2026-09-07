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
	JSONEntry struct {
		ID                     int                 `json:"id"`
		Name                   map[string]string   `json:"name,omitempty"`
		SimplifiedName         map[string]string   `json:"simplifiedName,omitempty"`
		Description            map[string]string   `json:"description,omitempty"`
		SimplifiedDescription  map[string]string   `json:"simplifiedDescription,omitempty"`
		Effect                 map[string]string   `json:"effect,omitempty"`
		Abilities              []map[string]string `json:"abilities,omitempty"`
		SensorText             map[string]string   `json:"sensorText,omitempty"`
		SimplifiedSensorText   map[string]string   `json:"simplifiedSensorText,omitempty"`
		ScanText               map[string]string   `json:"scanText,omitempty"`
		SimplifiedScanText     map[string]string   `json:"simplifiedScanText,omitempty"`
	}
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
		Description map[string]string   `json:"description"`
		Abilities   []map[string]string `json:"abilities"`
		Effect      map[string]string   `json:"effect"`
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
func updateObjectByType(jsonEntry JSONEntry, obj datastore.IGlobalLocalizedTextObject) error {
	switch localizedTextObj := obj.(type) {
	case *NameDescriptionTextObject:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
		updateSimplifiedNameEntry(jsonEntry, localizedTextObj.SimplifiedName)
		updateDescriptionEntry(jsonEntry, localizedTextObj.Description)
		updateSimplifiedDescriptionEntry(jsonEntry, localizedTextObj.SimplifiedDescription)
	case *NameDescriptionTextObjectV2:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
		updateDescriptionEntry(jsonEntry, localizedTextObj.Description)
	case *NameOnlyTextObject:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
	case *NameOnlyTextObjectV2:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
	case *DescriptionOnlyTextObject:
		updateDescriptionEntry(jsonEntry, localizedTextObj.Description)
		updateSimplifiedDescriptionEntry(jsonEntry, localizedTextObj.SimplifiedDescription)
	case *NameDescriptionEffect:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
		updateDescriptionEntry(jsonEntry, localizedTextObj.Description)
		updateEffectEntry(jsonEntry, localizedTextObj.Effect)
	case *NameDescriptionEffectAbilityTextObject:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
		updateDescriptionEntry(jsonEntry, localizedTextObj.Description)
		updateAbilitiesEntry(jsonEntry, localizedTextObj.Abilities)
	case *NameSensorScanTextObject:
		updateNameEntry(jsonEntry, localizedTextObj.Name)
		updateSensorTextEntry(jsonEntry, localizedTextObj)
		updateScanTextEntry(jsonEntry, localizedTextObj)
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
//   - itemsData: Slice of JSONEntry from JSON file
//   - objects: Slice of ILocalizedTextObject instances to be updated
//
// Returns: error if any update operations fail
func updateLocalizedObjectEntries(itemsData []JSONEntry, objects []datastore.IGlobalLocalizedTextObject) error {
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
//   - segment: The object to update
func updateNameEntry(sourceData JSONEntry, segment datastore.IGlobalLocalizedKeyedStringObject) {
	if len(sourceData.Name) == 0 || segment == nil {
		common.LogVerbose("No name found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, newNameText := range sourceData.Name {
		if newNameText == "" {
			common.LogVerbose("Empty name for language %s, ignoring...", languageCode)
			continue
		}
		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for name: %s", languageCode)
			continue
		}
		if newNameText == segment.GetLocalizedString(languageCode) {
			continue
		}

		updateOrCreateSegment(segment, newNameText, languageCode)
	}
}

// updateDescriptionEntry applies description updates to a NameDescriptionTextObject.
//
// Parameters:
//   - sourceData: JSON data containing description translations
//   - segment: The object to update
func updateDescriptionEntry(sourceData JSONEntry, segment datastore.IGlobalLocalizedKeyedStringObject) {
	if len(sourceData.Description) == 0 || segment == nil {
		common.LogVerbose("No description found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, newDescriptionText := range sourceData.Description {
		if newDescriptionText == "" {
			common.LogVerbose("Empty description for language %s, ignoring...", languageCode)
			continue
		}
		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for description: %s", languageCode)
			continue
		}
		if newDescriptionText == segment.GetLocalizedString(languageCode) {
			continue
		}

		updateOrCreateSegment(segment, newDescriptionText, languageCode)
	}
}

// updateSimplifiedNameEntry applies simplifiedName updates only if present in the JSON data.
//
// Parameters:
//   - sourceData: JSON data containing simplifiedName translations
//   - segment: The object to update
func updateSimplifiedNameEntry(sourceData JSONEntry, segment datastore.IGlobalLocalizedKeyedStringObject) {
	if len(sourceData.SimplifiedName) == 0 || segment == nil {
		return
	}

	for languageCode, newText := range sourceData.SimplifiedName {
		if newText == "" {
			continue
		}
		if !common.IsSupportedLanguage(languageCode) {
			continue
		}
		if newText == segment.GetLocalizedString(languageCode) {
			continue
		}

		updateOrCreateSegment(segment, newText, languageCode)
	}
}

// updateSimplifiedDescriptionEntry applies simplifiedDescription updates only if present in the JSON data.
//
// Parameters:
//   - sourceData: JSON data containing simplifiedDescription translations
//   - segment: The object to update
func updateSimplifiedDescriptionEntry(sourceData JSONEntry, segment datastore.IGlobalLocalizedKeyedStringObject) {
	if len(sourceData.SimplifiedDescription) == 0 || segment == nil {
		return
	}

	for languageCode, newText := range sourceData.SimplifiedDescription {
		if newText == "" {
			continue
		}
		if !common.IsSupportedLanguage(languageCode) {
			continue
		}
		if newText == segment.GetLocalizedString(languageCode) {
			continue
		}

		updateOrCreateSegment(segment, newText, languageCode)
	}
}

func updateAbilitiesEntry(sourceData JSONEntry, segments []datastore.IGlobalLocalizedKeyedStringObject) {
	if len(sourceData.Abilities) == 0 {
		common.LogVerbose("No abilities found for item %d, skipping...", sourceData.ID)
		return
	}

	// Limita o número de abilities para o que o objeto suporta
	maxAbilities := len(segments)
	if maxAbilities == 0 {
		common.LogVerbose("Target object has no abilities slots for item %d", sourceData.ID)
		return
	}

	for abilityIdx, abilityMap := range sourceData.Abilities {
		// Pula se não houver dados ou se exceder o número de abilities do objeto
		if len(abilityMap) == 0 {
			continue
		}
		if abilityIdx >= maxAbilities {
			common.LogVerbose("Ability %d exceeds target capacity (%d), skipping...",
				abilityIdx+1, maxAbilities)
			continue
		}

		segment := segments[abilityIdx]
		if segment == nil {
			common.LogVerbose("Ability %d segment is nil, skipping...", abilityIdx+1)
			continue
		}

		for languageCode, newAbilityText := range abilityMap {
			if newAbilityText == "" {
				continue
			}
			if !common.IsSupportedLanguage(languageCode) {
				continue
			}
			if newAbilityText == segment.GetLocalizedString(languageCode) {
				continue
			}

			updateOrCreateSegment(segment, newAbilityText, languageCode)
			common.LogVerbose("Updated ability %d for language %s: %s",
				abilityIdx+1, languageCode, newAbilityText)
		}
	}
}

// updateSensorTextEntry applies sensor text updates to a NameSensorScanTextObject.
//
// Parameters:
//   - sourceData: JSON data containing sensor translations
//   - targetObject: The NameSensorScanTextObject object to update
func updateSensorTextEntry(sourceData JSONEntry, targetObject *NameSensorScanTextObject) {
	if len(sourceData.SensorText) > 0 && targetObject.SensorText != nil {
		for languageCode, newText := range sourceData.SensorText {
			if newText == "" {
				continue
			}
			if !common.IsSupportedLanguage(languageCode) {
				continue
			}
			if newText == targetObject.SensorText.GetLocalizedString(languageCode) {
				continue
			}
			updateOrCreateSegment(targetObject.SensorText, newText, languageCode)
		}
	}

	if len(sourceData.SimplifiedSensorText) > 0 && targetObject.SimplifiedSensorText != nil {
		for languageCode, newText := range sourceData.SimplifiedSensorText {
			if newText == "" {
				continue
			}
			if !common.IsSupportedLanguage(languageCode) {
				continue
			}
			if newText == targetObject.SimplifiedSensorText.GetLocalizedString(languageCode) {
				continue
			}
			updateOrCreateSegment(targetObject.SimplifiedSensorText, newText, languageCode)
		}
	}
}

// updateScanTextEntry applies scan text updates to a NameSensorScanTextObject.
//
// Parameters:
//   - sourceData: JSON data containing scan translations
//   - targetObject: The NameSensorScanTextObject object to update
func updateScanTextEntry(sourceData JSONEntry, targetObject *NameSensorScanTextObject) {
	if len(sourceData.ScanText) > 0 && targetObject.ScanText != nil {
		for languageCode, newText := range sourceData.ScanText {
			if newText == "" {
				continue
			}
			if !common.IsSupportedLanguage(languageCode) {
				continue
			}
			if newText == targetObject.ScanText.GetLocalizedString(languageCode) {
				continue
			}
			updateOrCreateSegment(targetObject.ScanText, newText, languageCode)
		}
	}

	if len(sourceData.SimplifiedScanText) > 0 && targetObject.SimplifiedScanText != nil {
		for languageCode, newText := range sourceData.SimplifiedScanText {
			if newText == "" {
				continue
			}
			if !common.IsSupportedLanguage(languageCode) {
				continue
			}
			if newText == targetObject.SimplifiedScanText.GetLocalizedString(languageCode) {
				continue
			}
			updateOrCreateSegment(targetObject.SimplifiedScanText, newText, languageCode)
		}
	}
}

// updateEffectEntry applies effect updates to a NameDescriptionEffect.
func updateEffectEntry(sourceData JSONEntry, segment datastore.IGlobalLocalizedKeyedStringObject) {
	if len(sourceData.Effect) == 0 || segment == nil {
		common.LogVerbose("No effect found for item %d, skipping...", sourceData.ID)
		return
	}

	for languageCode, newEffectText := range sourceData.Effect {
		if newEffectText == "" {
			common.LogVerbose("Empty effect for language %s, ignoring...", languageCode)
			continue
		}

		if !common.IsSupportedLanguage(languageCode) {
			common.LogVerbose("Not recognized location for effect: %s", languageCode)
			continue
		}

		if newEffectText == segment.GetLocalizedString(languageCode) {
			continue
		}

		updateOrCreateSegment(segment, newEffectText, languageCode)
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
//   - text: The new text content
//   - languageCode: Language code (e.g., "us", "sp")
func updateOrCreateSegment(segment datastore.IGlobalLocalizedKeyedStringObject, newText, languageCode string) {
	existingContent := segment.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(newText, charset)
	} else {
		segment.SetLocalizedContent(languageCode, createNewKeyedString(newText, charset))
	}

	common.LogVerbose("Segment updated (%s): %s", languageCode, newText)
}

// updatePlateEntry applies plate updates to a PlateTextObject.
func updatePlateEntry(sourceData PlateTextData, targetObject *NameDescriptionEffectAbilityTextObject) {
	if len(sourceData.Name) > 0 && targetObject.Name != nil {
		for languageCode, nameText := range sourceData.Name {
			if nameText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.Name, nameText, languageCode)
		}
	}

	if len(sourceData.Description) > 0 && targetObject.Description != nil {
		for languageCode, descText := range sourceData.Description {
			if descText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.Description, descText, languageCode)
		}
	}

	for i, abMap := range sourceData.Abilities {
		if i >= len(targetObject.Abilities) {
			break
		}
		for languageCode, abilityText := range abMap {
			if abilityText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.Abilities[i], abilityText, languageCode)
		}
	}

	if len(sourceData.Effect) > 0 && targetObject.Effect != nil {
		for languageCode, effectText := range sourceData.Effect {
			if effectText == "" {
				continue
			}
			updateOrCreateSegment(targetObject.Effect, effectText, languageCode)
		}
	}
}

// serializePlateTextToJSON serializes a NameDescriptionEffectAbilityTextObject to JSON-friendly data.
// Returns a PlateTextData struct with name, description, abilities, and effect for all supported languages.
func serializePlateTextToJSON(obj *NameDescriptionEffectAbilityTextObject, languageCode string) PlateTextData {
	name := obj.GetName(languageCode)
	description := ""
	if descContent := obj.Description.GetLocalizedContent(languageCode); descContent != nil {
		description = descContent.GetString()
	}

	effect := ""
	if effContent := obj.Effect.GetLocalizedContent(languageCode); effContent != nil {
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
		/* Abilities: []map[string]string{
			{languageCode: ability},
		}, */
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

		plateObj, ok := objects[id].(*NameDescriptionEffectAbilityTextObject)
		if !ok {
			continue
		}

		updatePlateEntry(jsonEntry, plateObj)
	}

	common.LogVerbose("Plate objects updated from JSON: %s", jsonFileName)
	return nil
}
