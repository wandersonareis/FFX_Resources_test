package reader

import (
	"ffxresources/backend/common"
	"ffxresources/backend/fileFormats/event"
	"ffxresources/backend/interactions"
)

// ReadAllEvents reads all event files from the specified events directory.
// This function is a wrapper around the refactored converter.ReadAllEventFiles function
// to maintain backward compatibility with existing code.
//
// Parameters:
//   - eventsFolder: FileAccessor pointing to the root events directory
//
// Returns: error if the operation fails
func ReadAllEvents(eventsFolder common.FileAccessor) error {
	version := interactions.NewInteractionService().FFXAppConfig().GetGameVersion()
	return event.ReadAllEventFiles(eventsFolder, common.ToInt(version))
}

// ReadEventFull reads a complete event file with all localizations.
// This function is a wrapper around the refactored converter.ReadCompleteEventFile function
// to maintain backward compatibility with existing code.
//
// Parameters:
//   - eventId: Unique identifier for the event
//
// Returns: Complete EventFile with all localizations, or error if loading fails
/* func ReadEventFull(eventId string) (*components.EventFile, error) {
	return converter.ReadCompleteEventFile(eventId)
} */

// ReadEventFile reads a single event binary file and creates an EventFile object.
// This function is a wrapper around the refactored converter.ReadEventBinaryFile function
// to maintain backward compatibility with existing code.
//
// Parameters:
//   - eventId: Unique identifier for the event
//   - path: FileAccessor pointing to the event binary file
//
// Returns: EventFile object with binary data loaded, or error if reading fails
/* func ReadEventFile(eventId string, path common.FileAccessor) (*components.EventFile, error) {
	return converter.ReadEventBinaryFile(eventId, path)
} */
