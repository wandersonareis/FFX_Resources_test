package formats

import "ffxresources/backend/lib"

func getDialogsHandler(targetExtension ...string) (string, error) {
	extension := lib.DEFAULT_APPLICATION_EXTENSION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		lib.DEFAULT_RESOURCES_ROOTDIR,
		lib.DIALOG_HANDLER_APPLICATION,
	}

	tempProvider := lib.NewInteraction().TempProvider.ProvideTempFileWithExtension(lib.DIALOG_HANDLER_APPLICATION, extension)

	targetFile := tempProvider.FilePath
	err := lib.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
