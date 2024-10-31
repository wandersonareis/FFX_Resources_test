package formats

import "ffxresources/backend/common"

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
	extension := common.DEFAULT_APPLICATION_EXTENSION

	if len(targetExtension) > 0 {
		extension = targetExtension[0]
	}

	targetHandler := []string{
		common.DEFAULT_RESOURCES_ROOTDIR,
		common.DCP_FILE_XPLITTER_APPLICATION,
	}

	targetFile := common.NewTempProvider().ProvideTempFileWithExtension(common.DCP_FILE_XPLITTER_APPLICATION, extension).FilePath

	err := common.GetFileFromResources(targetHandler, targetFile)
	if err != nil {
		return "", err
	}

	return targetFile, nil
}
