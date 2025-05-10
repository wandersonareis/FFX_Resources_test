package lib

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

/*
CreateTemp creates a temporary working environment by setting up a temporary file and updating the source and destination accordingly.

Parameters:
	- source: An implementation of interfaces.ISource used to represent the original data source.
	- destination: An implementation of locations.IDestination used for file operations, including import and extract actions.

Returns:
	- The source (interfaces.ISource) with its path updated to the destination's import target file.
	- A new destination (locations.IDestination) whose extract location is modified to point to the temporary file and its associated path, while maintaining its translate and import locations.

This function leverages a temporary provider to generate a temporary file and adjusts the provided source and destination objects for further processing.
*/
func CreateTemp(source interfaces.ISource, destination locations.IDestination) (interfaces.ISource, locations.IDestination) {
	tmp := common.NewTempProvider("tmp", ".txt")

	tmpSource := source
	tmpSource.SetPath(destination.Import().GetTargetFile())

	extractLocation := destination.Extract().Copy()
	extractLocation.SetTargetFile(tmp.TempFile)
	extractLocation.SetTargetPath(tmp.TempFilePath)

	tmpDestination := &locations.Destination{
		ExtractLocation:   &extractLocation,
		TranslateLocation: destination.Translate(),
		ImportLocation:    destination.Import(),
	}

	return tmpSource, tmpDestination
}