package internal

import "ffxresources/backend/fileFormats/lib"

func getLockitFileHandler() (string, error) {
	targetFile, err := lib.GetFromResources(lib.LOCKIT_HANDLER_APPLICATION, lib.DEFAULT_APPLICATION_EXTENSION)
	if err != nil {
		return "", err
	}
	
	/* extension := common.DEFAULT_APPLICATION_EXTENSION
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
	} */

	return targetFile, nil
}

func getLockitFileUtf8BomNormalizer() (string, error) {
	targetFile, err := lib.GetFromResources(lib.UTF8BOM_NORMALIZER_APPLICATION, lib.DEFAULT_APPLICATION_EXTENSION)
	if err != nil {
		return "", err
	}

	/* extension := common.DEFAULT_APPLICATION_EXTENSION
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
	} */

	return targetFile, nil
}
