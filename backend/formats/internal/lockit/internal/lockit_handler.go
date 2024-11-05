package lockit_internal

import "ffxresources/backend/common"

func getLockitFileHandler(targetExtension ...string) (string, error) {
	extension := common.DEFAULT_APPLICATION_EXTENSION
	handlerApp := common.LOCKIT_HANDLER_APPLICATION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		handlerApp,
	}

	tempProvider := common.NewTempProvider()
	tempProvide := tempProvider.ProvideTempFileWithExtension(handlerApp, extension)

	targetFile := tempProvide.FilePath

	if err := common.GetFileFromResources(targetHandler, targetFile); err != nil {
		return "", err
	}

	return targetFile, nil
}

func getLockitFileUtf8BomNormalizer() (string, error) {
	extension := common.DEFAULT_APPLICATION_EXTENSION
	handlerApp := common.UTF8BOM_NORMALIZER_APPLICATION

	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		handlerApp,
	}

	tempProvider := common.NewTempProvider()
	tempProvide := tempProvider.ProvideTempFileWithExtension(handlerApp, extension)

	targetFile := tempProvide.FilePath

	if err := common.GetFileFromResources(targetHandler, targetFile); err != nil {
		return "", err
	}

	return targetFile, nil
}
