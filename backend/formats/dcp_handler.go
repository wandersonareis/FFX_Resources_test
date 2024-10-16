package formats

import "ffxresources/backend/lib"

/* func getDcpFileXpliter() (string, error) {
	exeName := "bin/SHSplit.exe"
	return lib.GetToolExcutable(exeName)
} */

/* func getDcpFileXpliterDev() (string, error) {
	handlerPath, err := lib.GetDcpXplitHandler()
	if err != nil {
		return "", err
	}
	return exec.LookPath(handlerPath)
} */

func GetDcpXplitHandler(targetExtension ...string) (string, error) {
	extension := lib.DEFAULT_APPLICATION_EXTENSION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		lib.DEFAULT_RESOURCES_ROOTDIR,
		lib.DCP_FILE_XPLITTER_APPLICATION,
	}

	tempProvider := lib.NewInteraction().TempProvider.ProvideTempFileWithExtension(lib.DCP_FILE_XPLITTER_APPLICATION, extension)

	targetFile := tempProvider.FilePath

	err := lib.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
