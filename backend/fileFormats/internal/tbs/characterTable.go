package tbstables

import "ffxresources/backend/fileFormats/lib"

type CharacterTable struct{}

func NewCharacterTable() *CharacterTable {
	return &CharacterTable{}
}

func (t *CharacterTable) GetFfx2CharacterTable() (string, error) {
	targetFile, err := lib.GetFromResources(lib.FFX2_CODE_TABLE_NAME, lib.DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (t *CharacterTable) GetCharacterOnlyTable() (string, error) {
	targetFile, err := lib.GetFromResources(lib.CHARACTER_CODE_TABLE, lib.DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (t *CharacterTable) GetCharacterLocTable() (string, error) {
	targetFile, err := lib.GetFromResources(lib.CHARACTER_LOC_CODE_TABLE, lib.DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

/* func getTableFromResources(codeTableName string) (string, error) {
	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		codeTableName,
	}

	tempProvider := common.NewTempProvider()
	tempProvide := tempProvider.ProvideTempFileWithExtension(codeTableName, "tbs")

	targetFile := tempProvide.FilePath

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
} */
