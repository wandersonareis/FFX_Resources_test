package internal

import (
	"ffxresources/backend/fileFormats/util"
	"os"
)

type lockitHandler struct {
	targetFile string
}

func newLockitHandler() *lockitHandler {
	return &lockitHandler{}
}

func (lh *lockitHandler) getLockitFileHandler() (string, error) {
	targetFile, err := util.GetFromResources(util.UTILS_RESOURCES_DIR, util.LOCKIT_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	lh.targetFile = targetFile

	return lh.targetFile, nil
}

func (lh *lockitHandler) getLockitFileUtf8BomNormalizer() (string, error) {
	targetFile, err := util.GetFromResources(util.UTILS_RESOURCES_DIR, util.UTF8BOM_NORMALIZER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	lh.targetFile = targetFile

	return lh.targetFile, nil
}

func (lh *lockitHandler) dispose() {
	if lh.targetFile != "" {
		os.Remove(lh.targetFile)
		lh.targetFile = ""
	}
}
