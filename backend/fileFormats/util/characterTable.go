package util

import "os"

type CharacterTable struct{}

func NewCharacterTable() *CharacterTable {
	return &CharacterTable{}
}

func (ct *CharacterTable) GetFfx2CharacterTable() (string, error) {
	targetFile, err := GetFromResources(CHARACTER_TABLE_RESOURCES_DIR, FFX2_ENCODING_TABLE_NAME, DEFAULT_ENCODING_TABLE_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (ct *CharacterTable) GetCharacterOnlyTable() (string, error) {
	targetFile, err := GetFromResources(CHARACTER_TABLE_RESOURCES_DIR, CHARACTER_ENCODING_TABLE, DEFAULT_ENCODING_TABLE_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (ct *CharacterTable) GetCharacterLocTable() (string, error) {
	targetFile, err := GetFromResources("tbs", CHARACTER_LOC_ENCODING_TABLE, DEFAULT_ENCODING_TABLE_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}

func (ct *CharacterTable) Dispose(file string) {
	os.Remove(file)
}
