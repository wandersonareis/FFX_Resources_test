package fileFormat

import "ffxresources/backend/lib"

func getCharacterTable() (string, error) {
	tablePath, err := lib.GetCharacterTable()
	if err != nil {
		return "", err
	}

	return tablePath, nil
}