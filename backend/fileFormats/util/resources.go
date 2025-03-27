package util

import "ffxresources/backend/common"

func GetFromResources(subDir, resourceName, ext string) (string, error) {
	targetHandler := []string{
		subDir,
		resourceName,
	}

	tempProvider := common.NewTempProvider(resourceName, ext)

	targetFile := tempProvider.TempFile

	if common.IsFileExists(targetFile) {
		return targetFile, nil
	}

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
