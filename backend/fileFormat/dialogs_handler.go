package fileFormat

import (
	"ffxresources/backend/lib"
	"os/exec"
)

func getDialogsFileHandler() (string, error) {
	handlerPath, err := lib.GetDialogsHandler()
	if err != nil {
		return "", err
	}

	return exec.LookPath(handlerPath)
}