package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
)

func getKernelFileHandler(targetExtension ...string) (string, error) {
	extension := common.DEFAULT_APPLICATION_EXTENSION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		common.KERNEL_HANDLER_APPLICATION,
	}

	tempProvider := lib.NewInteraction().TempProvider.ProvideTempFileWithExtension(common.KERNEL_HANDLER_APPLICATION, extension)

	targetFile := tempProvider.FilePath
	err := lib.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
