package dcp_internal

import "ffxresources/backend/common"

func GetDcpXplitHandler(targetExtension ...string) (string, error) {
	extension := common.DEFAULT_APPLICATION_EXTENSION
	handlerApp := common.DCP_FILE_XPLITTER_APPLICATION

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
