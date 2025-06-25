package converter

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/components"
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

	// EventFileData represents an event file with all its strings for JSON processing
	EventFileData struct {
		ID      string            `json:"id"`
		Strings []EventStringData `json:"strings"`
	}

	// EventStringData represents a single event string with its localizations for JSON processing
	EventStringData struct {
		Index int               `json:"index"`
		Text  map[string]string `json:"text"`
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
	objectsList components.IList[components.ILocalizedTextObject],
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

// ImportEventsDataFromJsonFile imports event data from a JSON file and applies
// the translations to all events found in the JSON file.
//
// This function reads JSON files containing event string information for all
// supported languages and applies the translations directly to the global EVENTS
// variable. Each event contains multiple strings with localized text content.
//
// File format: JSON array with event id and strings array containing localizations
// JSON structure: [{"id": "ev001", "strings": [{"index": 0, "text": {"us": "...", "sp": "..."}}]}]
//
// JSON file: events_all_localizations.json
// Target: components.EVENTS (multiple entries)
//
// Returns: error if import fails or file cannot be read
func ImportEventsDataFromJsonFile() error {
	return importEventJsonFile("", false)
}

// ImportEventDataFromJsonFile imports event data from a JSON file and applies
// the translations to a single specified event.
//
// This function reads JSON files containing event string information and applies
// the translations to a specific event in the global EVENTS variable. The event
// is identified by its ID and must exist in the JSON file.
//
// JSON file: events_all_localizations.json
// Target: components.EVENTS (single entry specified by eventID)
//
// Parameters:
//   - eventID: The ID of the specific event to process (e.g., "ev001", "btl_001")
//
// Returns: error if import fails, file cannot be read, or event is not found
func ImportEventDataFromJsonFile(eventID string) error {
	return importEventJsonFile(eventID, true)
}

// importEventJsonFile is the core function that handles both single and multiple event processing.
// This function centralizes the common logic between ImportEventsDataFromJsonFile and ImportEventDataFromJsonFile
// to avoid code duplication while providing flexibility for different processing modes.
//
// Parameters:
//   - eventID: The specific event ID to process (empty string for all events)
//   - singleEvent: Whether to process only a single event (true) or all events (false)
//
// Returns: error if processing fails
func importEventJsonFile(eventID string, singleEvent bool) error {
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

// getEventsJsonFilePath constructs and validates the path to the events JSON file.
// This function handles the standard path construction and existence validation.
//
// Returns: validated file path string, or error if file doesn't exist
func getEventsJsonFilePath() (string, error) {
	jsonPath := filepath.Join(common.GameFilesRoot, common.ModsFolder, "edits")
	if !common.IsPathExists(jsonPath) {
		return "", fmt.Errorf("edits directory not found: %s", jsonPath)
	}

	jsonFilePath := filepath.Join(jsonPath, "events_all_localizations.json")
	if !common.IsPathExists(jsonFilePath) {
		return "", fmt.Errorf("events JSON file not found: %s", jsonFilePath)
	}

	return jsonFilePath, nil
}

// loadEventJsonData loads and parses the events JSON file into EventFileData structures.
// This function handles file reading, JSON parsing, and basic validation.
//
// Parameters:
//   - jsonFilePath: Path to the events JSON file
//
// Returns: slice of EventFileData structures, or error if loading/parsing fails
func loadEventJsonData(jsonFilePath string) ([]EventFileData, error) {
	common.LogVerbose("Loading events JSON file: %s", jsonFilePath)

	eventDataList, err := common.ReadJsonFile[[]EventFileData](jsonFilePath)
	if err != nil {
		common.LogVerbose("Error reading events JSON file %s: %v", jsonFilePath, err)
		return nil, fmt.Errorf("failed to read events JSON file: %w", err)
	}

	common.LogVerbose("Successfully loaded %d events from JSON", len(eventDataList))
	return eventDataList, nil
}

// processSingleEventData processes a single event from the JSON data.
// This function finds the specified event in the data and applies its changes.
//
// Parameters:
//   - eventDataList: List of all event data from JSON
//   - eventID: The specific event ID to process
//
// Returns: error if event is not found or processing fails
func processSingleEventData(eventDataList []EventFileData, eventID string) error {
	common.LogVerbose("Looking for specific event: %s", eventID)

	var targetEventData *EventFileData
	for i := range eventDataList {
		if eventDataList[i].ID == eventID {
			targetEventData = &eventDataList[i]
			break
		}
	}

	if targetEventData == nil {
		return fmt.Errorf("event %s not found in JSON file", eventID)
	}

	common.LogVerbose("Event %s found in JSON with %d strings", eventID, len(targetEventData.Strings))

	if err := updateEventFromJsonData(*targetEventData); err != nil {
		return fmt.Errorf("failed to update event %s: %w", eventID, err)
	}

	if err := ExportEventStringsToLocalizations(eventID); err != nil {
		common.LogVerbose("Error saving event %s: %v", eventID, err)
		return fmt.Errorf("failed to save event %s: %w", eventID, err)
	}

	common.LogVerbose("Event %s processed and saved successfully", eventID)
	return nil
}

// processAllEventData processes all events from the JSON data.
// This function iterates through all events and applies their changes.
//
// Parameters:
//   - eventDataList: List of all event data from JSON
//
// Returns: error if any critical processing fails
func processAllEventData(eventDataList []EventFileData) error {
	common.LogVerbose("Processing all events from JSON (%d total)", len(eventDataList))

	processedEventIDs := make(map[string]bool)

	for _, eventData := range eventDataList {
		if err := updateEventFromJsonData(eventData); err != nil {
			common.LogVerbose("failed to update event %s: %v", eventData.ID, err)
			continue
		}
		processedEventIDs[eventData.ID] = true
	}

	common.LogVerbose("Events processed successfully! (%d events)", len(processedEventIDs))
	return nil
}

// updateEventFromJsonData updates a single event's data based on JSON input.
// This function handles the core logic of applying JSON changes to event strings.
//
// Parameters:
//   - eventData: The event data from JSON containing updates to apply
//
// Returns: error if the event is not found in memory or updating fails
func updateEventFromJsonData(eventData EventFileData) error {
	eventFile, exists := components.EVENTS[eventData.ID]
	if !exists || eventFile == nil {
		return fmt.Errorf("event not found in memory: %s", eventData.ID)
	}

	common.LogVerbose("Processing event %s with %d strings", eventData.ID, len(eventData.Strings))

	for _, eventString := range eventData.Strings {
		if err := updateEventStringFromJson(eventFile, eventString, eventData.ID); err != nil {
			common.LogVerbose("failed to update event %s: %v", eventData.ID, err)
			continue
		}
	}

	return nil
}

// updateEventStringFromJson updates a single string within an event based on JSON data.
// This function handles the string-level updates including localization processing.
//
// Parameters:
//   - eventFile: The event file object to update
//   - eventString: The string data from JSON
//   - eventID: The event ID (for logging purposes)
//
// Returns: error if string index is invalid or update fails
func updateEventStringFromJson(eventFile *components.EventFile, eventString EventStringData, eventID string) error {
	stringIndex := eventString.Index

	common.LogVerbose("Processing string %d for event %s", stringIndex, eventID)

	if stringIndex < 0 || stringIndex >= len(eventFile.Strings) {
		return fmt.Errorf("string index out of range for event %s: %d", eventID, stringIndex)
	}

	objToEdit := eventFile.Strings[stringIndex]

	common.LogVerbose("Updating event %s[%d] with %d localizations",
		eventID, stringIndex, len(eventString.Text))

	for localization, newString := range eventString.Text {
		if err := updateEventStringLocalization(objToEdit, localization, newString); err != nil {
			common.LogVerbose("failed to update localization %s for event %s[%d]: %v",
				localization, eventID, stringIndex, err)
		}
	}

	return nil
}

// updateEventStringLocalization updates a single localization for an event string object.
// This function handles the low-level localization update logic for event strings.
//
// Parameters:
//   - objToEdit: The string object to update
//   - localization: The localization code (e.g., "us", "jp")
//   - newString: The new string content
//
// Returns: error if localization is not supported or update fails
func updateEventStringLocalization(objToEdit *components.LocalizedFieldStringObject, localization, newString string) error {
	if newString == "" {
		return nil
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

// extractLocalizedObjects extracts ILocalizedTextObject instances from an IList
// and returns them as a regular slice for processing.
//
// Parameters:
//   - objectsList: IList containing ILocalizedTextObject entries
//
// Returns: Slice of ILocalizedTextObject instances, or nil if the list is empty
func extractLocalizedObjects(objectsList components.IList[components.ILocalizedTextObject]) []components.ILocalizedTextObject {
	if objectsList == nil || objectsList.IsEmpty() {
		return nil
	}

	return objectsList.GetItems()
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
func updateObjectByType(jsonEntry NameDescriptionData, obj components.ILocalizedTextObject) error {
	switch typed := obj.(type) {
	case *components.NameDescriptionTextObject:
		updateNameEntry(jsonEntry, typed)
		updateDescriptionEntry(jsonEntry, typed)
	case *components.NameOnlyTextObject:
		updateNameOnlyFromObjectsData(jsonEntry, typed)
	case *components.NameOnlyDataObject:
		nameOnly := typed.GetNameOnlyTextObject()
		if nameOnly != nil {
			updateNameOnlyFromObjectsData(jsonEntry, nameOnly)
		}
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
func updateLocalizedObjectEntries(itemsData []NameDescriptionData, objects []components.ILocalizedTextObject) error {
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
func updateNameEntry(sourceData NameDescriptionData, targetObject *components.NameDescriptionTextObject) {
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
func updateDescriptionEntry(sourceData NameDescriptionData, targetObject *components.NameDescriptionTextObject) {
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
func updateNameOnlyFromObjectsData(sourceData NameDescriptionData, targetObject *components.NameOnlyTextObject) {
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
func createNewKeyedString(text string, charset string) *components.KeyedString {
	return &components.KeyedString{
		Charset: charset,
		Offset:  0,
		Key:     0,
		Bytes:   components.StringToBytes(text, charset),
	}
}

// updateOrCreateName updates or creates a name entry for a NameDescriptionTextObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateName(target *components.NameDescriptionTextObject, languageCode, newText string) {
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
func updateOrCreateDescription(target *components.NameDescriptionTextObject, languageCode, text string) {
	existingContent := target.Description.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(text, charset)
	} else {
		target.Description.SetLocalizedContent(languageCode, createNewKeyedString(text, charset))
	}

	common.LogVerbose("Description updated (%s): %s", languageCode, text)
}

// updateOrCreateNameOnly updates or creates a name entry for a NameOnlyTextObject.
//
// Parameters:
//   - target: The object to update
//   - languageCode: Language code (e.g., "us", "sp")
//   - text: The new text content
func updateOrCreateNameOnly(target *components.NameOnlyTextObject, languageCode, text string) {
	existingContent := target.Name.GetLocalizedContent(languageCode)
	charset := common.LanguageCodeToCharset(languageCode)

	if existingContent != nil {
		existingContent.SetString(text, charset)
	} else {
		target.Name.SetLocalizedContent(languageCode, createNewKeyedString(text, charset))
	}

	common.LogVerbose("Name updated (%s): %s", languageCode, text)
}
