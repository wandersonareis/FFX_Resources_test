package formats

import "ffxresources/backend/lib"

func getCharacterTable() (string, error) {
	targetHandler := []string{
		lib.DEFAULT_RESOURCES_ROOTDIR,
		lib.FFX_CODE_TABLE_NAME,
	}

	tempProvider := lib.NewInteraction().TempProvider.ProvideTempFile(lib.FFX_CODE_TABLE_NAME)

	targetFile := tempProvider.FilePath

	err := lib.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
