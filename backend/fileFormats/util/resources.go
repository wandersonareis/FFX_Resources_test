package util

import "ffxresources/backend/common"

func GetFromResources(resourceName, ext string) (string, error) {
	targetHandler := []string{
		DEFAULT_RESOURCES_ROOTDIR,
		resourceName,
	}

	tempProvider := common.NewTempProvider()
	tempProvide := tempProvider.ProvideTempFileWithExtension(resourceName, ext)

	targetFile := tempProvide.FilePath

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
