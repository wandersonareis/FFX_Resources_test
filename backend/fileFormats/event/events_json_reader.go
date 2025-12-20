package event

// ProcessEventsJsonFile reads a JSON file and updates all events data in EVENTS.
//
// This function imports event data from "events_all_localizations.json" and applies
// the translations to all events found in the JSON file. The JSON file should contain
// event string information for all supported languages.
//
// JSON file: events_all_localizations.json
// Target: EVENTS (multiple entries)
//
// Returns: error if import fails or file cannot be read
func ProcessEventsJsonFile() error {
	return ImportEventsDataFromJsonFile()
}

// ProcessEventJsonFile reads a JSON file and updates a specific event in EVENTS.
//
// This function imports event data from "events_all_localizations.json" and applies
// the translations to a single specified event. The JSON file should contain
// event string information for all supported languages.
//
// JSON file: events_all_localizations.json
// Target: EVENTS (single entry specified by eventID)
//
// Parameters:
//   - eventID: The ID of the specific event to process (e.g., "ev001", "btl_001")
//
// Returns: error if import fails, file cannot be read, or event is not found
func ProcessEventJsonFile(eventID string) error {
	return ImportEventDataFromJsonFile(eventID)
}
