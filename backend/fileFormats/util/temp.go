package util

import (
	"ffxresources/backend/common"
	"ffxresources/backend/core/locations"
	"ffxresources/backend/interactions"
	"ffxresources/backend/interfaces"
)

func CreateTemporaryFileInfo(filePath string, formatter interfaces.ITextFormatterDev) (interfaces.ISource, locations.IDestination, string) {
	tmpDir := common.NewTempProviderDev("", "").TempFilePath

	gamePart := interactions.NewInteractionService().FFXGameVersion().GetGameVersion()

	source, err := locations.NewSource(filePath, gamePart)
	if err != nil {
		return nil, nil, ""
	}

	destination := locations.NewDestination()

	destination.InitializeLocations(source, formatter)

	destination.Extract().Get().SetTargetPath(tmpDir)

	return source, destination, tmpDir
}
