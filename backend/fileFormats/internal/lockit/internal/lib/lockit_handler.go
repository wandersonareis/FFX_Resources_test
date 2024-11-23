package lib

import (
	"ffxresources/backend/fileFormats/util"
	"fmt"
	"os"
)

type lockitHandler struct {
	util.Checksum
	targetFile string
}

func newLockitHandler() *lockitHandler {
	return &lockitHandler{
		Checksum: util.Checksum{},
	}
}

func (lh *lockitHandler) getLockitFileHandler() (string, error) {
	targetFile, err := util.GetFromResources(LOCKIT_RESOURCES_DIR, LOCKIT_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	lh.SetChecksumString(LOCKIT_HANDLER_SHA256)

	if !lh.IsValid(targetFile) {
		return "", fmt.Errorf("invalid checksum for lockit file handler")
	}

	lh.targetFile = targetFile

	return lh.targetFile, nil
}

func (lh *lockitHandler) getLockitFileUtf8BomNormalizer() (string, error) {
	targetFile, err := util.GetFromResources(LOCKIT_RESOURCES_DIR, UTF8BOM_NORMALIZER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	lh.SetChecksumString(UTF8BOM_NORMALIZER_SHA256)

	if !lh.IsValid(targetFile) {
		return "", fmt.Errorf("invalid checksum for lockit file utf8bom normalizer")
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
