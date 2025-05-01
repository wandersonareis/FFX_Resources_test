package util

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interfaces"
)

func CreateTemporaryFileInfo(filePath string, formatter interfaces.ITextFormatter) (interfaces.ISource, locations.IDestination) {
	tmpDir := common.NewTempProvider("", "").TempFilePath

	source, err := locations.NewSource(filePath)
	if err != nil {
		return nil, nil
	}

	destination := locations.NewDestination()

	destination.InitializeLocations(source, formatter)

	destination.Extract().SetTargetPath(tmpDir)

	return source, destination
}
