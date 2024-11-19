package util

import "ffxresources/backend/common"

func GetFromResources(subDir, resourceName, ext string) (string, error) {
	targetHandler := []string{
		subDir,
		resourceName,
	}

	tempProvider := common.NewTempProvider()
	tempProvide := tempProvider.ProvideTempFileWithExtension(resourceName, ext)

	targetFile := tempProvide.File

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
