package internal

import "ffxresources/backend/formats/lib"

func getDialogsHandler() (string, error) {
	targetFile, err := lib.GetFromResources(lib.DIALOG_HANDLER_APPLICATION, lib.DEFAULT_TABLE_EXTENSION)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
