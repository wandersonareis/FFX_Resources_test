package encodingHandler

import (
	"ffxresources/backend/fileFormats/util"
	"fmt"
	"os"
)

type ILockitEncodingHandler interface {
	FetchLockitHandler() (string, error)
	FetchLockitUtf8BomNormalizer() (string, error)
	Dispose()
}

type lockitEncodingHandler struct {
	util.Checksum
	handlerFile string
	utf8BomFile string
}

func NewLockitHandler() *lockitEncodingHandler {
	return &lockitEncodingHandler{
		Checksum: util.Checksum{},
	}
}

func (lh *lockitEncodingHandler) FetchLockitHandler() (string, error) {
	targetFile, err := util.GetFromResources(LOCKIT_RESOURCES_DIR, LOCKIT_HANDLER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	if !lh.VerifyChecksum(targetFile, LOCKIT_HANDLER_SHA256) {
		return "", fmt.Errorf("invalid checksum for lockit file handler")
	}

	lh.handlerFile = targetFile

	return lh.handlerFile, nil
}

func (lh *lockitEncodingHandler) FetchLockitUtf8BomNormalizer() (string, error) {
	targetFile, err := util.GetFromResources(LOCKIT_RESOURCES_DIR, UTF8BOM_NORMALIZER_APPLICATION, util.DEFAULT_APPLICATION_FILE_EXTENSION)
	if err != nil {
		return "", err
	}

	if !lh.VerifyChecksum(targetFile, UTF8BOM_NORMALIZER_SHA256) {
		return "", fmt.Errorf("invalid checksum for lockit file utf8bom normalizer")
	}

	lh.utf8BomFile = targetFile

	return lh.utf8BomFile, nil
}

func (lh *lockitEncodingHandler) Dispose() {
	if lh.handlerFile != "" {
		if err := os.Remove(lh.handlerFile); err != nil {
			fmt.Println("error when removing lockit file handler")
		}

		lh.handlerFile = ""
	}

	if lh.utf8BomFile != "" {
		if err := os.Remove(lh.utf8BomFile); err != nil {
			fmt.Println("error when removing lockit file utf8bom normalizer")
		}
		
		lh.utf8BomFile = ""
	}
}
