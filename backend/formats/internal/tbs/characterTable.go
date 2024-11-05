package tbstables

import "ffxresources/backend/common"

type CharacterTable struct {}

func NewCharacterTable() *CharacterTable {
	return &CharacterTable{}
}

func (t *CharacterTable) GetFfx2CharacterTable() (string, error) {
	targetFile, err := getTableFromResources(common.FFX2_CODE_TABLE_NAME)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (t *CharacterTable) GetCharacterOnlyTable() (string, error) {
	targetFile, err := getTableFromResources(common.CHARACTER_CODE_TABLE)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (t *CharacterTable) GetCharacterLocTable() (string, error) {
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