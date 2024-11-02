package formats

import "ffxresources/backend/common"

func getFfx2CharacterTable() (string, error) {
	targetFile, err := getTableFromResources(common.FFX2_CODE_TABLE_NAME)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func getCharacterOnlyTable() (string, error) {
	targetFile, err := getTableFromResources(common.CHARACTER_CODE_TABLE)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func getCharacterLocTable() (string, error) {
	targetFile, err := getTableFromResources(common.CHARACTER_LOC_CODE_TABLE)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func getTableFromResources(codeTableName string) (string, error) {
	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		codeTableName,
	}

	targetFile := common.NewTempProvider().ProvideTempFileWithExtension(codeTableName, "tbs").FilePath

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}