package formats

import "ffxresources/backend/common"

func getKernelFileHandler(targetExtension ...string) (string, error) {
	extension := common.DEFAULT_APPLICATION_EXTENSION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		common.KERNEL_HANDLER_APPLICATION,
	}

	targetFile := common.NewTempProvider().ProvideTempFileWithExtension(common.KERNEL_HANDLER_APPLICATION, extension).FilePath

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
