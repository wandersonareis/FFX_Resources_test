package util

import "os"

type CharacterTable struct {
	targetFile string
}

func (ct *CharacterTable) GetFfx2CharacterTable() (string, error) {
	targetFile, err := GetFromResources(FFX2_CODE_TABLE_NAME, DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	ct.targetFile = targetFile

	return ct.targetFile, nil
}

func (ct *CharacterTable) GetCharacterOnlyTable() (string, error) {
	targetFile, err := GetFromResources(CHARACTER_CODE_TABLE, DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	ct.targetFile = targetFile

	return ct.targetFile, nil
}

func (ct *CharacterTable) GetCharacterLocTable() (string, error) {
	targetFile, err := GetFromResources(CHARACTER_LOC_CODE_TABLE, DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	ct.targetFile = targetFile

	return ct.targetFile, nil
}

func (ct *CharacterTable) Dispose() {
	if ct.targetFile != "" {
		os.Remove(ct.targetFile)
		ct.targetFile = ""
	}
}
