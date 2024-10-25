package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
)

func getCharacterTable() (string, error) {
	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		common.FFX_CODE_TABLE_NAME,
	}

	tempProvider := lib.NewInteraction().TempProvider.ProvideTempFile(common.FFX_CODE_TABLE_NAME)

	targetFile := tempProvider.FilePath

	err := lib.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
