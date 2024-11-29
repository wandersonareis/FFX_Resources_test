package util

import "ffxresources/backend/common"

func GetFromResources(subDir, resourceName, ext string) (string, error) {
	targetHandler := []string{
		subDir,
		resourceName,
	}

	tempProvider := common.NewTempProviderDev(resourceName, ext)

	targetFile := tempProvider.TempFile

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
