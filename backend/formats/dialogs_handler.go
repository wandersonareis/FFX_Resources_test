package formats

import (
	"ffxresources/backend/common"
	"ffxresources/backend/lib"
)

func getDialogsHandler(targetExtension ...string) (string, error) {
	extension := common.DEFAULT_APPLICATION_EXTENSION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		common.DIALOG_HANDLER_APPLICATION,
	}

	tempProvider := lib.NewInteraction().TempProvider.ProvideTempFileWithExtension(common.DIALOG_HANDLER_APPLICATION, extension)

	targetFile := tempProvider.FilePath
	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
